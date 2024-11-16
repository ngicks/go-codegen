package undgen

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/imports"
)

type UndParser struct {
	// Parser applies its modified behavior only for files under dir.
	dir  string
	mode parser.Mode
}

func NewUndParser(dir string) *UndParser {
	if !filepath.IsAbs(dir) {
		var err error
		dir, err = filepath.Abs(dir)
		if err != nil {
			panic(err)
		}
	}
	return &UndParser{
		dir:  dir,
		mode: parser.AllErrors | parser.ParseComments,
	}
}

func (p UndParser) ParseFile(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
	f, err := parser.ParseFile(fset, filename, src, p.mode)
	if err != nil {
		return f, err
	}

	if rel, err := filepath.Rel(p.dir, filename); err != nil || strings.Contains(rel, "..") {
		return f, err
	}

	f.Decls = slices.AppendSeq(
		f.Decls[:0],
		xiter.Filter(
			func(decl ast.Decl) bool {
				var (
					direction UndDirection
					ok        bool
					err       error
				)
				switch x := decl.(type) {
				case *ast.FuncDecl:
					direction, ok, err = ParseUndComment(x.Doc)
				case *ast.GenDecl:
					direction, ok, err = ParseUndComment(x.Doc)
					if direction.generated {
						return false
					}
					x.Specs = slices.AppendSeq(
						x.Specs[:0],
						xiter.Filter(
							func(spec ast.Spec) bool {
								direction, ok, err := ParseUndComment(x.Doc)
								if !ok || err != nil {
									// no error at this moment
									return true
								}
								return !direction.generated
							},
							slices.Values(x.Specs)),
					)
				}
				if !ok || err != nil {
					// no error at this moment
					return true
				}
				return !direction.generated
			},
			slices.Values(f.Decls),
		),
	)

	// Now, import decls might be used by any code inside the file.
	// Since we've removed some of them.
	//
	// To fix this situation, call golang.org/x/tools/imports.Process to remove unused imports.

	buf := new(bytes.Buffer)
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
