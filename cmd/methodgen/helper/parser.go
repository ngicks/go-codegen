package helper

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"strings"
	"go/types"
)

const ignoreCommentSlash = "//methodgen:ignore="

type Parser struct {
	PkgName string
	Info    map[string]TypeInfo
	Debug   bool
}

func (p *Parser) ParseDir(dirname string, targetTypeNames []string, ignoreComment []string) error {
	p.Info = map[string]TypeInfo{}

	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(
		fset,
		dirname,
		func(fi fs.FileInfo) bool {
			return fi.Mode().IsRegular() && !strings.HasSuffix(fi.Name(), "_test.go")
		},
		parser.ParseComments,
	)

	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		p.PkgName = pkg.Name
		v := &visitor{
			Parser:              p,
			TargetTypeNames:     targetTypeNames,
			IgnoreCommentSuffix: ignoreComment,
		}
		ast.Walk(v, pkg)
	}

	return nil
}

type visitor struct {
	*Parser
	TargetTypeNames     []string // def.DeIterator, def.SeIterator
	IgnoreCommentSuffix []string
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case types.Object:
		if v.Debug {
			fmt.Fprintf(os.Stderr, "package: %s\n", n.Name)
		}
		return v
	case *ast.File:
		if v.Debug {
			fmt.Fprintf(os.Stderr, "filestart, fileend = %d, %d\n", n.FileStart, n.FileEnd)
		}
		if hasIgnore(n.Doc, v.IgnoreCommentSuffix) {
			fmt.Fprintf(os.Stderr, "  *ignore: file level\n")
			return nil
		}
		return v
	case *ast.GenDecl:
		if hasIgnore(n.Doc, v.IgnoreCommentSuffix) {
			for _, nc := range n.Specs {
				switch nct := nc.(type) {
				case *ast.TypeSpec:
					nct.Doc = n.Doc
				}
			}
		}
		return v
	case *ast.TypeSpec:
		if v.Debug {
			fmt.Fprintf(os.Stderr, "  type name = %s\n", n.Name)
		}
		if hasIgnore(n.Doc, v.IgnoreCommentSuffix) {
			if v.Debug {
				out := make([]string, 0)
				for _, l := range n.Doc.List {
					out = append(out, l.Text)
				}

				fmt.Fprintf(os.Stderr, "  *ignored with %s\n", strings.Join(out, " "))
			}
			return nil
		}

		structTy, ok := n.Type.(*ast.StructType)
		if !ok {
			return v
		}

		for _, field := range structTy.Fields.List {
			idx, ok := field.Type.(*ast.IndexExpr)
			if !ok {
				continue
			}

			var typName string
			if selector, ok := idx.X.(*ast.SelectorExpr); ok {
				qual, ok := selector.X.(*ast.Ident)
				if !ok {
					continue
				}
				typName = qual.Name + "." + selector.Sel.Name
			} else {
				ident, ok := idx.X.(*ast.Ident)
				if !ok {
					continue
				}
				typName = ident.Name
			}

			if sliceHas(v.TargetTypeNames, typName) {
				var memberName string
				if len(field.Names) > 0 {
					memberName = field.Names[0].Name
				} else {
					// embedded
					if idx := strings.Index(typName, "."); idx >= 0 {
						memberName = typName[idx+1:]
					} else {
						memberName = typName
					}
				}
				v.Info[n.Name.Name] = TypeInfo{
					TypeName:        n.Name.Name,
					TypeParams:      getTypeParam(n),
					InnerMemberName: memberName,
					StructType:      structTy,
				}
			}
		}
	}
	return nil
}

func hasIgnore(comments *ast.CommentGroup, ignore []string) bool {
	if comments == nil {
		return false
	}

	for _, v := range comments.List {
		comment := v.Text
		if idx := strings.Index(comment, ignoreCommentSlash); idx < 0 {
			continue
		} else {
			comment = comment[idx+len(ignoreCommentSlash):]
		}

		for _, c := range strings.Split(strings.TrimSpace(comment), ",") {
			if c == "all" {
				return true
			}
			for _, ign := range ignore {
				if c == ign {
					return true
				}
			}
		}
	}

	return false
}

func sliceHas(tab []string, target string) bool {
	for _, t := range tab {
		if target == t {
			return true
		}
	}
	return false
}

// getTypeParam gets type parameters of typeSpec.
func getTypeParam(spec *ast.TypeSpec) []string {
	if spec.TypeParams == nil || spec.TypeParams.List == nil {
		return []string{}
	}

	tyParams := make([]string, 0)
	for _, v := range spec.TypeParams.List {
		for _, n := range v.Names {
			tyParams = append(tyParams, n.Name)
		}
	}
	return tyParams
}
