// parser defines general utilities for codegen.
package astutil

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ngicks/go-codegen/codegen/internal/bufpool"
	"github.com/ngicks/go-codegen/codegen/pkg/directive"
	"github.com/ngicks/go-iterator-helper/hiter"
	"golang.org/x/tools/imports"
)

type Parser struct {
	// Parser applies its modified behavior only for files under dir.
	dir  string
	mode parser.Mode
}

func NewParser(dir string) *Parser {
	if !filepath.IsAbs(dir) {
		var err error
		dir, err = filepath.Abs(dir)
		if err != nil {
			panic(err)
		}
	}
	return &Parser{
		dir:  dir,
		mode: parser.AllErrors | parser.ParseComments,
	}
}

func (p *Parser) ParseFile(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
	f, err := parser.ParseFile(fset, filename, src, p.mode)
	if err != nil {
		return f, err
	}

	if rel, err := filepath.Rel(p.dir, filename); err != nil || strings.Contains(rel, "..") {
		return f, err
	}

	var removedNodeRanges []tokenRange
	f.Decls = slices.AppendSeq(
		f.Decls[:0],
		hiter.Filter(
			func(decl ast.Decl) (pass bool) {
				var tokRange tokenRange
				defer func() {
					if pass || len(tokRange.filter()) == 0 {
						return
					}
					removedNodeRanges = append(removedNodeRanges, tokRange.filter())
				}()

				var (
					direction directive.Direction
					ok        bool
					err       error
				)
				switch x := decl.(type) {
				case *ast.FuncDecl:
					direction, ok, err = directive.ParseDirectiveComment(x.Doc)
					tokRange = append(tokRange, getCommentGroupPos(x.Doc), x.Pos(), x.End())
				case *ast.GenDecl:
					direction, ok, err = directive.ParseDirectiveComment(x.Doc)
					tokRange = append(tokRange, getCommentGroupPos(x.Doc), x.Pos(), x.End())
					if direction.IsGenerated() {
						return false
					}
					x.Specs = slices.AppendSeq(
						x.Specs[:0],
						hiter.Filter(
							func(spec ast.Spec) (pass bool) {
								var tokRange tokenRange
								defer func() {
									if pass || len(tokRange.filter()) == 0 {
										return
									}
									removedNodeRanges = append(removedNodeRanges, tokRange.filter())
								}()

								var (
									direction directive.Direction
									ok        bool
									err       error
								)
								switch x := spec.(type) { // IMPORT, CONST, TYPE, or VAR
								default:
									return true
								case *ast.ValueSpec:
									direction, ok, err = directive.ParseDirectiveComment(x.Comment)
									tokRange = append(tokRange, getCommentGroupPos(x.Doc), x.Pos(), x.End())
								case *ast.TypeSpec:
									direction, ok, err = directive.ParseDirectiveComment(x.Comment)
									tokRange = append(tokRange, getCommentGroupPos(x.Doc), x.Pos(), x.End())
								}
								if !ok || err != nil {
									// no error at this moment
									return true
								}
								return !direction.IsGenerated()
							},
							slices.Values(x.Specs)),
					)
				}
				if !ok || err != nil {
					// no error at this moment
					return true
				}
				return !direction.IsGenerated()
			},
			slices.Values(f.Decls),
		),
	)

	if len(removedNodeRanges) == 0 {
		return f, nil
	}

	f.Comments = slices.AppendSeq(
		f.Comments[:0],
		hiter.Filter(
			func(cg *ast.CommentGroup) bool {
				return !slices.ContainsFunc(
					removedNodeRanges,
					func(tokRange tokenRange) bool {
						return tokRange.IsWithin(cg.Pos())
					},
				)
			},
			slices.Values(f.Comments),
		),
	)

	// Now, import decls might not be used by any code inside the file.
	// Since we've removed some of them.
	//
	// To fix this situation, call golang.org/x/tools/imports.Process to remove unused imports.

	buf := bufpool.GetBuf()
	defer bufpool.PutBuf(buf)

	err = printer.Fprint(buf, fset, f)
	if err != nil {
		panic(err)
	}

	fixed, err := imports.Process(filename, buf.Bytes(), nil)
	if err != nil {
		panic(err)
	}

	return parser.ParseFile(fset, filename, fixed, parser.AllErrors|parser.ParseComments)
}

type tokenRange []token.Pos

func (r tokenRange) filter() tokenRange {
	return slices.AppendSeq(
		r[:0],
		hiter.Filter(
			func(pos token.Pos) bool { return pos != 0 },
			slices.Values(r)),
	)
}

func (r tokenRange) IsWithin(pos token.Pos) bool {
	if len(r) < 2 {
		panic("tokenRange: invalid range")
	}
	return r[0] <= pos && pos <= r[len(r)-1]
}

func getCommentGroupPos(cg *ast.CommentGroup) token.Pos {
	if cg == nil || len(cg.List) == 0 { // invariants disallow len(cg.List) == 0 but check it anyway.
		return 0
	}
	return cg.Pos()
}