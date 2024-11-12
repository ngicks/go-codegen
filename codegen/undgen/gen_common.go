package undgen

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"reflect"
	"slices"

	"github.com/dave/dst"
	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-codegen/codegen/structtag"
	"github.com/ngicks/go-iterator-helper/hiter"
)

func or[T any](left bool, l, r T) T {
	if left {
		return l
	} else {
		return r
	}
}

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

// returns conversion.Empty
func conversionEmptyExpr(importMap imports.ImportMap) *dst.SelectorExpr {
	return importMap.DstExpr(UndTargetTypeConversionEmpty)
}

func bufPrintf(w io.Writer) (func(format string, args ...any), func() error) {
	bufw := bufio.NewWriter(w)
	return func(format string, args ...any) {
			fmt.Fprintf(bufw, format, args...)
		}, func() error {
			return bufw.Flush()
		}
}

func importIdent(ty imports.TargetType, imports imports.ImportMap) string {
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

func fieldJsonName(st *types.Struct, i int) string {
	tags, _ := structtag.ParseStructTag(reflect.StructTag(st.Tag(i)))
	if _, name, err := tags.Get("json", ""); err == nil {
		return name
	}
	return st.Field(i).Name()
}

func makeRenamedType(ty *types.Named, name string, pkg *types.Package, method func(typeName *types.TypeName) []*types.Func) *types.Named {
	obj := types.NewTypeName(0, pkg, name, nil)
	funs := method(obj)
	renamed := types.NewNamed(obj, nil, funs)
	renamed.SetUnderlying(ty.Underlying())
	if ty.TypeArgs().Len() == 0 {
		return renamed
	}
	instantiated, err := types.Instantiate(
		nil,
		renamed,
		slices.Collect(hiter.OmitF(hiter.AtterAll(ty.TypeArgs()))),
		false,
	)
	if err != nil {
		panic(err)
	}
	aa := types.TypeString(instantiated, (*types.Package).Name)
	_ = aa
	return instantiated.(*types.Named)
}
