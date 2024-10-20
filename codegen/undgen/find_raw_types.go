package undgen

import (
	"go/ast"
	"go/types"
	"reflect"

	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/undtag"
	"golang.org/x/tools/go/packages"
)

// FindRawTypes checks types defined in pkgs
// if they include types listed imports or implementors of conversion methods described in methods.
// The types listed in the returned value is assumed to be later processed to
// become an implementor of [ConversionMethodsSet.ToPlain] named in methods.
//
// pkgs are the packages which contains the types to be checked.
// Callers can load it though `golang.org/x/tools/go/packages`.
// PkgPath, Fset, Syntax and TypesInfo fields of each [*packages.Package] will be used.
// Thus the loader must at least load with [packages.NeedName]|[packages.NeedSyntax]|[packages.NeedTypesInfo] bits.
//
// pkgs should only contain packages under cwd
// because the target types are normally code generation target;
// types based on the targets with some modification will be generated and
// written along the source codes which define those base types.
// FindRawTypes itself does not take any verification about path correctness.
// Basically callers should not sym-link any directories under the target.
//
// imports is the slice of package paths and type names.
// The defined struct/slice/array/map types
// which contain those listed in imports as field / element / value type are considered a target.
//
// methods declares the conversion methods between raw types,
// which this functions is trying to find, and plain types.
// The type is an implementor when [ConversionMethodsSet.ToPlain] is implemented on the type and its returned value is single.
// And the returned value also implements [ConversionMethodsSet.ToRaw] and its returned value is single and the type.
//
// FindRawTypes ignores types if they have //undgen:ignore or //undgen:generated in the associated doc comments.
func FindRawTypes(pkgs []*packages.Package, imports []TargetImport, methods ConversionMethodsSet) (RawTypes, error) {
	// 1st path, find other than implementor
	matched, err := findRawTypes(pkgs, imports, methods, nil)
	if err != nil {
		return matched, err
	}
	// 2nd path, find including implementor
	matched, err = findRawTypes(pkgs, imports, methods, matched)
	if err != nil {
		return matched, err
	}

	return matched, nil
}

type TargetImport struct {
	ImportPath string
	Types      []string
}

type ConversionMethodsSet struct {
	ToRaw   string
	ToPlain string
}

type RawTypes map[string]RawMatchedPackage

func (m RawTypes) HasTy(path string, name string) (RawMatchedType, bool) {
	pkg, ok := m[path]
	if !ok {
		return RawMatchedType{}, false
	}
	return pkg.HasTy(name)
}

type RawMatchedPackage struct {
	Pkg   *packages.Package
	Files map[string]RawMatchedFile
}

func (m RawMatchedPackage) HasTy(name string) (RawMatchedType, bool) {
	for _, f := range m.Files {
		if t, ok := f.HasTy(name); ok {
			return t, ok
		}
	}
	return RawMatchedType{}, false
}

func (mt *RawMatchedPackage) lazyInit(pkg *packages.Package) {
	if mt.Files == nil {
		mt.Pkg = pkg
		mt.Files = make(map[string]RawMatchedFile)
	}
}

type RawMatchedFile struct {
	File     *ast.File
	Filename string
	Types    map[int]RawMatchedType
}

func (mf *RawMatchedFile) lazyInit(f *ast.File, filename string) {
	if mf.Types == nil {
		mf.File = f
		mf.Filename = filename
		mf.Types = make(map[int]RawMatchedType)
	}
}

func (mf RawMatchedFile) HasTy(name string) (RawMatchedType, bool) {
	for _, t := range mf.Types {
		if t.Name == name {
			return t, true
		}
	}
	return RawMatchedType{}, false
}

// RawMatchedType indicates the type defined in a file is a target type.
//
// There should be 4 kind
type RawMatchedType struct {
	// 0-indexed number of appearance within the file. source code order.
	Pos int

	GenDecl  *ast.GenDecl
	TypeSpec *ast.TypeSpec
	TypeInfo types.Object

	// Name of type without type params.
	// Just here for later reuse to look up ast.
	Name string
	// this must not be MatchedAsImplementor
	Variant MatchedAs
	// len(Field) == 1 if Variants is other than "struct".
	Field []MatchedField
}

type MatchedField struct {
	Pos  int
	Name string
	As   MatchedAs
	// Empty if As is "implementor".
	Type TargetType
	// Elem type for "array", "slice", "map".
	// In that case Type should be "direct".
	Elem MatchedFieldElem
	Tag  option.Option[hiter.KeyValue[undtag.UndOpt, error]]
}

func (mf MatchedField) IsValid() bool {
	return mf.As != ""
}

type MatchedFieldElem struct {
	As   MatchedAs
	Type TargetType
}

type MatchedAs string

const (
	MatchedAsDirect      MatchedAs = "direct"
	MatchedAsStruct      MatchedAs = "struct"
	MatchedAsArray       MatchedAs = "array"
	MatchedAsSlice       MatchedAs = "slice"
	MatchedAsMap         MatchedAs = "map"
	MatchedAsImplementor MatchedAs = "implementor"
)

func findRawTypes(pkgs []*packages.Package, imports []TargetImport, methods ConversionMethodsSet, matched RawTypes) (RawTypes, error) {
	if matched == nil {
		matched = make(RawTypes)
	}

	for pkg, seq := range enumerateTypeSpec(pkgs) {
		matchedPkg := matched[pkg.PkgPath]
		matchedPkg.lazyInit(pkg)

		for file, seq := range seq {
			importMap := parseImports(file.Imports, imports)

			filename := pkg.Fset.Position(file.FileStart).Filename
			matchedFile := matchedPkg.Files[filename]
			matchedFile.lazyInit(file, filename)

			for tsi := range seq {
				if tsi.Err != nil {
					return matched, tsi.Err
				}

				mt, ok := parseUndType(tsi.TypeInfo, matched, importMap, methods)
				if !ok {
					continue
				}

				mt.Pos = tsi.Pos
				mt.GenDecl = tsi.GenDecl
				mt.TypeSpec = tsi.TypeSpec
				mt.TypeInfo = tsi.TypeInfo

				matchedFile.Types[mt.Pos] = mt
			}
			matchedPkg.Files[filename] = matchedFile
		}

		matched[pkg.PkgPath] = matchedPkg
	}
	return matched, nil
}

func parseUndType(
	obj types.Object,
	total RawTypes,
	imports importDecls,
	conversionMethod ConversionMethodsSet,
) (mt RawMatchedType, has bool) {
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return
	}
	switch underlying := obj.Type().Underlying().(type) {
	default:
		return RawMatchedType{}, false
	case *types.Struct:
		var matched []MatchedField
		for i := range underlying.NumFields() {
			f := underlying.Field(i)
			matchedAs := isRawType(f.Type(), imports, total, conversionMethod)
			if !matchedAs.IsValid() {
				continue
			}
			matchedAs.Pos = i
			matchedAs.Name = f.Name()
			matchedAs.Tag = option.MapOption(
				option.FromOk(reflect.StructTag(underlying.Tag(i)).Lookup(undtag.TagName)),
				func(tagLit string) hiter.KeyValue[undtag.UndOpt, error] {
					tag, err := undtag.ParseOption(tagLit)
					return hiter.KeyValue[undtag.UndOpt, error]{K: tag, V: err}
				},
			)
			matched = append(matched, matchedAs)
		}
		return RawMatchedType{
			Name:    named.Obj().Name(),
			Variant: MatchedAsStruct,
			Field:   matched,
		}, len(matched) > 0
	case *types.Array:
		matchedAs := isRawType(underlying.Elem(), imports, total, conversionMethod)
		return RawMatchedType{
			Name:    named.Obj().Name(),
			Variant: MatchedAsArray,
			Field:   []MatchedField{matchedAs},
		}, matchedAs.IsValid()
	case *types.Slice:
		matchedAs := isRawType(underlying.Elem(), imports, total, conversionMethod)
		return RawMatchedType{
			Name:    named.Obj().Name(),
			Variant: MatchedAsSlice,
			Field:   []MatchedField{matchedAs},
		}, matchedAs.IsValid()
	case *types.Map:
		matchedAs := isRawType(underlying.Elem(), imports, total, conversionMethod)
		return RawMatchedType{
			Name:    named.Obj().Name(),
			Variant: MatchedAsMap,
			Field:   []MatchedField{matchedAs},
		}, matchedAs.IsValid()
	}
}

func isRawType(
	ty types.Type,
	imports importDecls,
	total RawTypes,
	conversionMethod ConversionMethodsSet,
) (mf MatchedField) {
	switch x := ty.(type) {
	case *types.Named:
		pkgPath, name := x.Obj().Pkg().Path(), x.Obj().Name()
		targetTy, ok := imports.MatchTy(pkgPath, name)
		if ok {
			filed := MatchedField{
				As:   MatchedAsDirect,
				Type: targetTy,
			}
			typeArg := x.TypeArgs()
			if typeArg.Len() > 0 {
				tt := isRawType(typeArg.At(0), imports, total, conversionMethod)
				if tt.IsValid() {
					filed.Elem = MatchedFieldElem{As: tt.As, Type: tt.Type}
				}
			}
			return filed
		}
		_, ok = total.HasTy(pkgPath, name)
		if ok {
			return MatchedField{As: MatchedAsImplementor}
		}
		if isImplementor(x, conversionMethod, false) {
			return MatchedField{As: MatchedAsImplementor}
		}
	case *types.Array:
		m := isRawType(x.Elem(), imports, total, conversionMethod)
		if m.As != "" {
			return MatchedField{
				As: MatchedAsArray,
				Elem: MatchedFieldElem{
					As:   m.As,
					Type: m.Type,
				},
			}
		}
	case *types.Slice:
		m := isRawType(x.Elem(), imports, total, conversionMethod)
		if m.As != "" {
			return MatchedField{
				As: MatchedAsSlice,
				Elem: MatchedFieldElem{
					As:   m.As,
					Type: m.Type,
				},
			}
		}
	case *types.Map:
		m := isRawType(x.Elem(), imports, total, conversionMethod)
		if m.As != "" {
			return MatchedField{
				As: MatchedAsMap,
				Elem: MatchedFieldElem{
					As:   m.As,
					Type: m.Type,
				},
			}
		}
	}

	return MatchedField{}
}

// isImplementor checks if ty can be converted to a type, then converted back from the type to ty
// though methods described in conversionMethod.
//
// Assuming fromPlain is false, ty is an implementor if ty (called type A hereafter)
// has the method which [ConversionMethodsSet.ToPlain] names
// where the returned value of the method is only one and type B,
// and also type B implements the method which [ConversionMethodsSet.ToRaw] describes
// where the returned value of the method is only one and type A.
//
// If fromPlain is true isImplementor works reversely (it checks assuming ty is type B.)
func isImplementor(ty *types.Named, conversionMethod ConversionMethodsSet, fromPlain bool) bool {
	toMethod := conversionMethod.ToPlain
	revMethod := conversionMethod.ToRaw
	if fromPlain {
		toMethod, revMethod = revMethod, toMethod
	}

	ms := types.NewMethodSet(ty)
	for i := range ms.Len() {
		sel := ms.At(i)
		if sel.Obj().Name() == toMethod {
			sig, ok := sel.Obj().Type().Underlying().(*types.Signature)
			if !ok {
				return false
			}
			tup := sig.Results()
			if tup.Len() != 1 {
				return false
			}
			v := tup.At(0)

			named, ok := v.Type().(*types.Named)
			if !ok {
				return false
			}

			ms := types.NewMethodSet(named)
			for i := range ms.Len() {
				sel := ms.At(i)
				if sel.Obj().Name() == revMethod {
					sig, ok := sel.Obj().Type().Underlying().(*types.Signature)
					if !ok {
						return false
					}
					tup := sig.Results()
					if tup.Len() != 1 {
						return false
					}
					v := tup.At(0)

					named, ok := v.Type().(*types.Named)
					if !ok {
						return false
					}

					objStr1 := ty.Obj().String() // Assigning to a value just to inspect the string in the debugger.
					objStr2 := named.Obj().String()
					// simple pointer comparison should not suffice since
					// if types are instantiated, they can be same type but different pointer.
					// Am I correct? At least if I replace the line below with `return ty == named`
					// Test_isImplementor fails.
					return objStr1 == objStr2
				}
			}

		}
	}
	return false
}
