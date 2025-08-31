package astutil

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/ngicks/go-codegen/codegen/internal/bufpool"
	"github.com/ngicks/go-codegen/codegen/pkg/imports"
	"github.com/ngicks/go-codegen/codegen/pkg/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
)

// BufPrintf wraps w and returns it in a function which works as like printf.
// Call second return value to retrieve io error in case it happens.
func BufPrintf(w io.Writer) (func(format string, args ...any), func() error) {
	bufw := bufio.NewWriter(w)
	return func(format string, args ...any) {
			fmt.Fprintf(bufw, format, args...)
		}, func() error {
			return bufw.Flush()
		}
}

func PrintTypeParamsAst(ts *ast.TypeSpec) string {
	if ts.TypeParams == nil || len(ts.TypeParams.List) == 0 {
		return ""
	}
	var typeParams strings.Builder
	for _, f := range ts.TypeParams.List {
		for _, name := range f.Names {
			if typeParams.Len() > 0 {
				typeParams.WriteByte(',')
			}
			typeParams.WriteString(name.Name)
		}
	}
	return "[" + typeParams.String() + "]"
}

func PrintTypeParamsDst(ts *dst.TypeSpec) string {
	if ts.TypeParams == nil || len(ts.TypeParams.List) == 0 {
		return ""
	}
	var typeParams strings.Builder
	for _, f := range ts.TypeParams.List {
		for _, name := range f.Names {
			if typeParams.Len() > 0 {
				typeParams.WriteByte(',')
			}
			typeParams.WriteString(name.Name)
		}
	}
	return "[" + typeParams.String() + "]"
}

func PrintAstExprPanicking(expr ast.Expr) string {
	buf := bufpool.GetBuf()
	defer bufpool.PutBuf(buf)

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
				hiter.Map(
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
	default:
		panic(fmt.Errorf("unknown type: %T", ty))
	// This branch does not return value.
	// This might be parametrized. Later parameters are examined.
	case hasObj:
		if x.Obj() != nil &&
			x.Obj().Pkg() != nil &&
			x.Obj().Pkg().Path() != pkgPath {
			exp = importMap.AstExpr(imports.TargetType{
				ImportPath: x.Obj().Pkg().Path(),
				Name:       x.Obj().Name(),
			})
		} else {
			// Built-in named interfaces like error.
			exp = &ast.Ident{
				Name: x.Obj().Name(),
			}
		}
		// returns immediately.
	case hasName:
		// basic, const, var, func, label, etc.
		return &ast.Ident{
			Name: x.Name(),
		}
	case *types.Pointer:
		return &ast.StarExpr{
			X: TypeToAst(x.Elem(), pkgPath, importMap),
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
	case *types.Struct:
		fields := slices.Collect(
			hiter.Unify(
				func(i int, v *types.Var) *ast.Field {
					var tag *ast.BasicLit
					if tagLit := x.Tag(i); len(tagLit) > 0 {
						tag = &ast.BasicLit{
							Kind:  token.STRING,
							Value: strconv.Quote(tagLit),
						}
					}
					return &ast.Field{
						Names: []*ast.Ident{ast.NewIdent(v.Name())},
						Type:  TypeToAst(v.Type(), pkgPath, importMap),
						Tag:   tag,
					}
				},
				pkgsutil.EnumerateFields(x),
			),
		)
		var opening, closing token.Pos
		if len(fields) == 0 {
			opening = 1
			closing = 2
			// this will make printed tok looks like struct{}.
			// Prettier when struct{} is used as size 0 data.
		}
		return &ast.StructType{
			Fields: &ast.FieldList{
				Opening: opening,
				List:    fields,
				Closing: closing,
			},
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
				hiter.Map(
					func(ty types.Type) ast.Expr {
						return TypeToAst(ty, pkgPath, importMap)
					},
					hiter.OmitF(hiter.AtterAll(named.TypeArgs())),
				),
			),
		}
	}
}

func PrintFileHeader(w io.Writer, af *ast.File, fset *token.FileSet) error {
	if err := printPackage(w, af); err != nil {
		return err
	}
	if err := printImport(w, af, fset); err != nil {
		return err
	}
	return nil
}

func printPackage(w io.Writer, af *ast.File) error {
COMMENTS:
	for _, cg := range af.Comments {
		for _, c := range cg.List {
			if c.End() > af.Package {
				break COMMENTS
			}
			_, err := w.Write([]byte(c.Text))
			if err != nil {
				return err
			}
			if strings.HasPrefix(c.Text, "//") {
				_, err := w.Write([]byte("\n"))
				if err != nil {
					return err
				}
			}
		}
		_, err := w.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "%s %s\n\n",
		token.PACKAGE.String(), af.Name.Name,
	); err != nil {
		return err
	}
	return nil
}

func printImport(w io.Writer, af *ast.File, fset *token.FileSet) error {
	for i, dec := range af.Decls {
		genDecl, ok := dec.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.IMPORT {
			// it's possible that the file has multiple import spec.
			// but it always starts with import spec.
			break
		}
		err := printer.Fprint(w, fset, genDecl)
		if err != nil {
			return fmt.Errorf("print.Fprint failed printing %dth import spec: %w", i, err)
		}
		_, err = io.WriteString(w, "\n")
		if err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\n")
	if err != nil {
		return err
	}

	return nil
}

