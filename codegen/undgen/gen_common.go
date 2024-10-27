package undgen

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/ngicks/go-codegen/codegen/pkgsutil"
	"github.com/ngicks/go-iterator-helper/hiter"
)

func printPackage(w io.Writer, af *ast.File) error {
	if _, err := io.WriteString(w, (token.PACKAGE.String())); err != nil {
		return err
	}
	if _, err := w.Write([]byte{' '}); err != nil {
		return err
	}
	if _, err := io.WriteString(w, (af.Name.Name)); err != nil {
		return err
	}
	if _, err := io.WriteString(w, ("\n\n")); err != nil {
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
		_, err = io.WriteString(w, "\n\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// returns *struct{}
func startStructExpr() *dst.StarExpr {
	return &dst.StarExpr{
		X: &dst.StructType{
			Fields: &dst.FieldList{Opening: true, Closing: true},
		},
	}
}

func bufPrintf(w io.Writer) (func(format string, args ...any), func() error) {
	bufw := bufio.NewWriter(w)
	return func(format string, args ...any) {
			fmt.Fprintf(bufw, format, args...)
		}, func() error {
			return bufw.Flush()
		}
}

func printTypeParamForField(fset *token.FileSet, ts *ast.TypeSpec, fieldName string) (string, error) {
	st := ts.Type.(*ast.StructType) // panic if not a struct
	if st.Fields == nil {
		return "", fmt.Errorf("struct has no field")
	}
	for _, field := range st.Fields.List {
		if !slices.ContainsFunc(field.Names, func(n *ast.Ident) bool { return n.Name == fieldName }) {
			continue
		}
		var node ast.Node
		switch t := field.Type.(type) {
		default:
			return "", nil
		case *ast.IndexExpr:
			node = t // this includes field type itself; IndexExpr is X[Index]
		case *ast.IndexListExpr:
			node = t
		}
		buf := new(bytes.Buffer)
		err := printer.Fprint(buf, fset, node)
		if err != nil {
			return "", err
		}
		s := buf.String()
		// Instead of removing non-index part from node(it might interfere with other part of code)
		// just trim down output string.
		return s[strings.Index(s, "[")+1 : len(s)-1], nil
	}
	return "", fmt.Errorf("field not found: %q", fieldName)
}

func importIdent(ty TargetType, imports importDecls) string {
	optionImportIdent, _ := imports.Ident(UndTargetTypeOption.ImportPath)
	undImportIdent, _ := imports.Ident(UndTargetTypeUnd.ImportPath)
	sliceUndImportIdent, _ := imports.Ident(UndTargetTypeSliceUnd.ImportPath)
	elasticImportIdent, _ := imports.Ident(UndTargetTypeElastic.ImportPath)
	sliceElasticImportIdent, _ := imports.Ident(UndTargetTypeSliceElastic.ImportPath)
	switch ty {
	case UndTargetTypeElastic:
		return elasticImportIdent
	case UndTargetTypeSliceElastic:
		return sliceElasticImportIdent
	case UndTargetTypeUnd:
		return undImportIdent
	case UndTargetTypeSliceUnd:
		return sliceUndImportIdent
	case UndTargetTypeOption:
		return optionImportIdent
	}
	return ""
}

func appendTypeAndTypeParams(imports []TargetImport, pkgPath string, ty types.Type) []TargetImport {
	p, ok := ty.(*types.Pointer)
	if ok {
		ty = p.Elem()
	}
	named, ok := ty.(*types.Named)
	if !ok {
		return imports
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil {
		return imports
	}
	if pkg.Path() != "" && pkg.Path() != pkgPath {
		imports = AppendTargetImports(imports,
			TargetImport{
				ImportPath: pkg.Path(),
				Types:      []string{obj.Name()},
			},
		)
	}
	for ty := range pkgsutil.EnumerateTypeParams(named) {
		if ty == nil {
			continue
		}
		p, ok := ty.(*types.Pointer)
		if ok {
			ty = p.Elem()
		}
		named, ok := ty.(*types.Named)
		if !ok {
			continue
		}
		if named.Obj() == nil || named.Obj().Pkg() == nil {
			continue
		}
		if named.Obj().Pkg().Path() == pkgPath {
			continue
		}
		imports = AppendTargetImports(
			imports,
			TargetImport{
				ImportPath: named.Obj().Pkg().Path(),
				Types:      []string{named.Obj().Name()},
			},
		)
	}
	return imports
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

func typeToDst(ty types.Type, pkgPath string, importMap importDecls) dst.Expr {
	var exp dst.Expr
	switch x := ty.(type) {
	case *types.Pointer:
		exp = &dst.StarExpr{
			X: typeToDst(x.Elem(), pkgPath, importMap),
		}
	case hasName:
		exp = &dst.Ident{
			Name: x.Name(),
		}
	case hasObj:
		if x.Obj() != nil &&
			x.Obj().Pkg() != nil &&
			x.Obj().Pkg().Path() != pkgPath {
			exp = importMap.DstExpr(TargetType{
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
			Index: typeToDst(named.TypeArgs().At(0), pkgPath, importMap),
		}
	default:
		args := named.TypeArgs()
		var exprs []dst.Expr
		for _, ty := range hiter.IndexAccessible(args, hiter.Range(0, args.Len())) {
			exprs = append(exprs, typeToDst(ty, pkgPath, importMap))
		}
		return &dst.IndexListExpr{
			X:       exp,
			Indices: exprs,
		}
	}
}

func typeToAst(ty types.Type, pkgPath string, importMap importDecls) ast.Expr {
	var exp ast.Expr
	switch x := ty.(type) {
	case *types.Pointer:
		exp = &ast.StarExpr{
			X: typeToAst(x.Elem(), pkgPath, importMap),
		}
	case hasName:
		exp = &ast.Ident{
			Name: x.Name(),
		}
	case hasObj:
		if x.Obj() != nil &&
			x.Obj().Pkg() != nil &&
			x.Obj().Pkg().Path() != pkgPath {
			exp = importMap.AstExpr(TargetType{
				ImportPath: x.Obj().Pkg().Path(),
				Name:       x.Obj().Name(),
			})
		} else {
			exp = &ast.Ident{
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
		return &ast.IndexExpr{
			X:     exp,
			Index: typeToAst(named.TypeArgs().At(0), pkgPath, importMap),
		}
	default:
		args := named.TypeArgs()
		var exprs []ast.Expr
		for _, ty := range hiter.IndexAccessible(args, hiter.Range(0, args.Len())) {
			exprs = append(exprs, typeToAst(ty, pkgPath, importMap))
		}
		return &ast.IndexListExpr{
			X:       exp,
			Indices: exprs,
		}
	}
}
