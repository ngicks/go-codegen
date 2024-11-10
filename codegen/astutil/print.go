package astutil

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"slices"
	"strconv"

	"github.com/dave/dst"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
)

func PrintAstExprPanicking(expr ast.Expr) string {
	buf := new(bytes.Buffer)
	err := printer.Fprint(buf, token.NewFileSet(), expr)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type hasName interface {
	Name() string
}

type hasObj interface {
	Obj() *types.TypeName
}

type hasTypeArg interface {
	TypeArgs() *types.TypeList
}

func TypeToDst(ty types.Type, pkgPath string, importMap imports.ImportMap) dst.Expr {
	var exp dst.Expr
	switch x := ty.(type) {
	case *types.Pointer:
		exp = &dst.StarExpr{
			X: TypeToDst(x.Elem(), pkgPath, importMap),
		}
	case hasName:
		exp = &dst.Ident{
			Name: x.Name(),
		}
	case hasObj:
		if x.Obj() != nil &&
			x.Obj().Pkg() != nil &&
			x.Obj().Pkg().Path() != pkgPath {
			exp = importMap.DstExpr(imports.TargetType{
				ImportPath: x.Obj().Pkg().Path(),
				Name:       x.Obj().Name(),
			})
		} else {
			exp = &dst.Ident{
				Name: x.Obj().Name(),
			}
		}
	}

	named, ok := ty.(hasTypeArg)
	if !ok {
		return exp
	}
	switch named.TypeArgs().Len() {
	case 0:
		return exp
	case 1:
		return &dst.IndexExpr{
			X:     exp,
			Index: TypeToDst(named.TypeArgs().At(0), pkgPath, importMap),
		}
	default:
		return &dst.IndexListExpr{
			X: exp,
			Indices: slices.Collect(
				xiter.Map(
					func(ty types.Type) dst.Expr {
						return TypeToDst(ty, pkgPath, importMap)
					},
					hiter.OmitF(hiter.AtterAll(named.TypeArgs())),
				),
			),
		}
	}
}

func TypeToAst(ty types.Type, pkgPath string, importMap imports.ImportMap) ast.Expr {
	var exp ast.Expr
	switch x := ty.(type) {
	case *types.Pointer:
		exp = &ast.StarExpr{
			X: TypeToAst(x.Elem(), pkgPath, importMap),
		}
	case hasName:
		exp = &ast.Ident{
			Name: x.Name(),
		}
	case hasObj:
		if x.Obj() != nil &&
			x.Obj().Pkg() != nil &&
			x.Obj().Pkg().Path() != pkgPath {
			exp = importMap.AstExpr(imports.TargetType{
				ImportPath: x.Obj().Pkg().Path(),
				Name:       x.Obj().Name(),
			})
		} else {
			exp = &ast.Ident{
				Name: x.Obj().Name(),
			}
		}
	case *types.Array:
		return &ast.ArrayType{
			Len: &ast.BasicLit{
				Kind:  token.INT,
				Value: strconv.FormatInt(x.Len(), 10),
			},
			Elt: TypeToAst(x.Elem(), pkgPath, importMap),
		}
	case *types.Slice:
		return &ast.ArrayType{
			Elt: TypeToAst(x.Elem(), pkgPath, importMap),
		}
	case *types.Map:
		return &ast.MapType{
			Key:   TypeToAst(x.Key(), pkgPath, importMap),
			Value: TypeToAst(x.Elem(), pkgPath, importMap),
		}
	}

	named, ok := ty.(hasTypeArg)
	if !ok {
		return exp
	}
	switch named.TypeArgs().Len() {
	case 0:
		return exp
	case 1:
		return &ast.IndexExpr{
			X:     exp,
			Index: TypeToAst(named.TypeArgs().At(0), pkgPath, importMap),
		}
	default:
		return &ast.IndexListExpr{
			X: exp,
			Indices: slices.Collect(
				xiter.Map(
					func(ty types.Type) ast.Expr {
						return TypeToAst(ty, pkgPath, importMap)
					},
					hiter.OmitF(hiter.AtterAll(named.TypeArgs())),
				),
			),
		}
	}
}
