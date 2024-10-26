package undgen

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"slices"
	"strings"

	"github.com/dave/dst"
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
