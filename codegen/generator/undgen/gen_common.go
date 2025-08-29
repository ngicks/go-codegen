package undgen

import (
	"go/types"
	"reflect"
	"slices"

	"github.com/dave/dst"
	"github.com/ngicks/go-codegen/codegen/pkg/imports"
	"github.com/ngicks/go-codegen/codegen/pkg/structtag"
	"github.com/ngicks/go-iterator-helper/hiter"
)

func or[T any](left bool, l, r T) T {
	if left {
		return l
	} else {
		return r
	}
}

// returns conversion.Empty
func conversionEmptyExpr(importMap imports.ImportMap) *dst.SelectorExpr {
	return importMap.DstExpr(UndTargetTypeConversionEmpty)
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
	return instantiated.(*types.Named)
}
