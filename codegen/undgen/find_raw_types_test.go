package undgen

import (
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/ngicks/go-codegen/codegen/structtag"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/hiter/iterable"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/undtag"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

func Test_isImplementor(t *testing.T) {
	var foo, fooPlain, bar, nonCyclic *types.Named

	for _, pkg := range targettypesPackages {
		if pkg.PkgPath != "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub" {
			continue
		}
		for _, def := range pkg.TypesInfo.Defs {
			tn, ok := def.(*types.TypeName)
			if !ok {
				continue
			}
			n := tn.Name()
			named, _ := def.Type().(*types.Named)
			switch n {
			case "Foo":
				foo = named
			case "FooPlain":
				fooPlain = named
			case "Bar":
				bar = named
			case "NonCyclic":
				nonCyclic = named
			}
		}
	}

	assert.Assert(t, foo != nil)
	assert.Assert(t, fooPlain != nil)

	mset := ConversionMethodsSet{
		ToRaw:   "UndRaw",
		ToPlain: "UndPlain",
	}
	assertIsConversionMethodImplementor := func(ty *types.Named, conversionMethod ConversionMethodsSet, fromPlain bool, ok1, ok2 bool) {
		t.Helper()
		ty, ok_ := isConversionMethodImplementor(ty, conversionMethod, fromPlain)
		assert.Assert(t, ok1 == ok_)
		if ok2 {
			assert.Assert(t, ty != nil)
		} else {
			assert.Assert(t, ty == nil)
		}
	}

	assertIsConversionMethodImplementor(foo, mset, false, true, true)
	assertIsConversionMethodImplementor(fooPlain, mset, true, true, true)

	assertIsConversionMethodImplementor(bar, mset, true, false, false)
	assertIsConversionMethodImplementor(nonCyclic, mset, true, false, false)
}

func Test_parseImports(t *testing.T) {
	var file1, file2 *ast.File
P:
	for _, p := range targettypesPackages {
		for _, f := range p.Syntax {
			fPath := p.Fset.Position(f.FileStart)
			if strings.HasSuffix(fPath.Filename, "undgen/internal/targettypes/ty1.go") {
				file1 = f
			}
			if strings.HasSuffix(fPath.Filename, "undgen/internal/targettypes/ty2.go") {
				file2 = f
			}
			if file1 != nil && file2 != nil {
				break P
			}
		}
	}

	importMap := parseImports(file1.Imports, ConstUnd.Imports)
	expected := importDecls{
		identToImport: map[string]TargetImport{
			"option":   ConstUnd.Imports[0],
			"und":      ConstUnd.Imports[1],
			"elastic":  ConstUnd.Imports[2],
			"sliceund": ConstUnd.Imports[3],
		},
		missingImports: map[string]TargetImport{
			"elastic_1":  ConstUnd.Imports[4],
			"undtag":     ConstUnd.Imports[5],
			"validate":   ConstUnd.Imports[6],
			"conversion": ConstUnd.Imports[7],
		},
	}
	assert.DeepEqual(
		t,
		expected.identToImport,
		importMap.identToImport,
	)
	assert.DeepEqual(
		t,
		expected.missingImports,
		importMap.missingImports,
	)

	importMap = parseImports(file2.Imports, ConstUnd.Imports)
	expected = importDecls{
		identToImport: map[string]TargetImport{
			"option":       ConstUnd.Imports[0],
			"und":          ConstUnd.Imports[1],
			"elastic":      ConstUnd.Imports[2],
			"sliceund":     ConstUnd.Imports[3],
			"sliceElastic": ConstUnd.Imports[4],
		},
		missingImports: map[string]TargetImport{
			"undtag":     ConstUnd.Imports[5],
			"validate":   ConstUnd.Imports[6],
			"conversion": ConstUnd.Imports[7],
		},
	}
	assert.DeepEqual(
		t,
		expected.identToImport,
		importMap.identToImport,
	)
	assert.DeepEqual(
		t,
		expected.missingImports,
		importMap.missingImports,
	)
}

func TestFindTargetType_error(t *testing.T) {
	result, err := FindRawTypes(
		targettypesPackages,
		ConstUnd.Imports,
		ConstUnd.ConversionMethod,
		nil,
	)
	assert.Assert(t, err != nil)
	assert.Assert(t, len(result) > 0)
}

func deepEqualRawMatchedType(t *testing.T, i, j []hiter.KeyValue[int, RawMatchedType]) {
	t.Helper()
	assert.DeepEqual(
		t,
		i, j,
		// generally ignore additional fields.
		// other code / tests use these so isn't too bad to ignore here.
		gocmp.Comparer(func(i, j *ast.GenDecl) bool { return true }),
		gocmp.Comparer(func(i, j *ast.TypeSpec) bool { return true }),
		gocmp.Comparer(func(i, j types.Object) bool { return true }),
		gocmp.Comparer(func(i, j structtag.Tags) bool { return true }),
		gocmp.Comparer(func(i, j types.Type) bool { return true }),
	)
}

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}

var (
	tagRequired                = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("required"))})
	tagNullish                 = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("nullish"))})
	tagDef                     = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("def"))})
	tagNull                    = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("null"))})
	tagUnd                     = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("und"))})
	tagDefUnd                  = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("def,und"))})
	tagDefNull                 = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("def,null"))})
	tagNullUnd                 = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("null,und"))})
	tagDefNullUnd              = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("def,null,und"))})
	tagLenEq1                  = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("len==1"))})
	tagLenGt1                  = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("len>1"))})
	tagLenGte1                 = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("len>=1"))})
	tagLt1                     = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("len<1"))})
	tagLte1                    = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("len<=1"))})
	tagRequiredLenEq2          = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("required,len==2"))})
	tagNullishLenEq2           = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("nullish,len==2"))})
	tagDefLenEq2               = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("def,len==2"))})
	tagNullLenEq2              = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("null,len==2"))})
	tagUndLenEq2               = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("und,len==2"))})
	tagValuesNonNull           = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("values:nonnull"))})
	tagNullValuesNonNull       = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("null,values:nonnull"))})
	tagValuesNonNullLenEq1     = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("values:nonnull,len==1"))})
	tagNullValuesNonNullLenEq1 = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("null,values:nonnull,len==1"))})
	tagValuesNonNullEq3        = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("values:nonnull,len==3"))})
	tagNullValuesNonNullLenEq3 = option.Some(UndTagParseResult{Opt: must(undtag.ParseOption("null,values:nonnull,len==3"))})
)

func TestFindTargetType(t *testing.T) {
	result, err := FindRawTypes(
		slices.Collect(
			xiter.Filter(func(pkg *packages.Package) bool {
				return pkg.PkgPath != "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/erroneous"
			},
				slices.Values(targettypesPackages),
			),
		),
		ConstUnd.Imports,
		ConstUnd.ConversionMethod,
		nil,
	)
	assert.NilError(t, err)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pkg := result["github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes"]
	if pkg.Pkg.PkgPath != "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes" {
		t.Errorf("wrong path: %s", pkg.Pkg.PkgPath)
	}

	f := pkg.Files[filepath.Join(cwd, filepath.FromSlash("internal/targettypes/ty1.go"))]
	types := hiter.Collect2(iterable.MapSorted[int, RawMatchedType](f.Types).Iter2())
	deepEqualRawMatchedType(
		t,
		[]hiter.KeyValue[int, RawMatchedType]{
			{
				K: 0,
				V: RawMatchedType{
					Pos:     0,
					Name:    "All",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{Pos: 4, Name: "UntouchedOpt", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Pos: 5, Name: "UntouchedUnd", As: MatchedAsDirect, Type: UndTargetTypeUnd},
						{Pos: 6, Name: "UntouchedSliceUnd", As: MatchedAsDirect, Type: UndTargetTypeSliceUnd},
						{Pos: 7, Name: "OptRequired", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagRequired},
						{Pos: 8, Name: "OptNullish", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagNullish},
						{Pos: 9, Name: "OptDef", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagDef},
						{Pos: 10, Name: "OptNull", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagNull},
						{Pos: 11, Name: "OptUnd", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagUnd},
						{Pos: 12, Name: "OptDefOrUnd", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagDefUnd},
						{Pos: 13, Name: "OptDefOrNull", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagDefNull},
						{Pos: 14, Name: "OptNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagNullUnd},
						{Pos: 15, Name: "OptDefOrNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeOption, UndTag: tagDefNullUnd},
						{Pos: 16, Name: "UndRequired", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagRequired},
						{Pos: 17, Name: "UndNullish", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagNullish},
						{Pos: 18, Name: "UndDef", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagDef},
						{Pos: 19, Name: "UndNull", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagNull},
						{Pos: 20, Name: "UndUnd", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagUnd},
						{Pos: 21, Name: "UndDefOrUnd", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagDefUnd},
						{Pos: 22, Name: "UndDefOrNull", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagDefNull},
						{Pos: 23, Name: "UndNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagNullUnd},
						{Pos: 24, Name: "UndDefOrNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeUnd, UndTag: tagDefNullUnd},
						{Pos: 25, Name: "ElaRequired", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagRequired},
						{Pos: 26, Name: "ElaNullish", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagNullish},
						{Pos: 27, Name: "ElaDef", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagDef},
						{Pos: 28, Name: "ElaNull", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagNull},
						{Pos: 29, Name: "ElaUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagUnd},
						{Pos: 30, Name: "ElaDefOrUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagDefUnd},
						{Pos: 31, Name: "ElaDefOrNull", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagDefNull},
						{Pos: 32, Name: "ElaNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagNullUnd},
						{Pos: 33, Name: "ElaDefOrNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagDefNullUnd},
						{Pos: 34, Name: "ElaEqEq", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagLenEq1},
						{Pos: 35, Name: "ElaGr", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagLenGt1},
						{Pos: 36, Name: "ElaGrEq", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagLenGte1},
						{Pos: 37, Name: "ElaLe", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagLt1},
						{Pos: 38, Name: "ElaLeEq", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagLte1},
						{Pos: 39, Name: "ElaEqEquRequired", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagRequiredLenEq2},
						{Pos: 40, Name: "ElaEqEquNullish", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagNullishLenEq2},
						{Pos: 41, Name: "ElaEqEquDef", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagDefLenEq2},
						{Pos: 42, Name: "ElaEqEquNull", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagNullLenEq2},
						{Pos: 43, Name: "ElaEqEquUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagUndLenEq2},
						{Pos: 44, Name: "ElaEqEqNonNullSlice", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagValuesNonNull},
						{Pos: 45, Name: "ElaEqEqNonNullNullSlice", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagNullValuesNonNull},
						{Pos: 46, Name: "ElaEqEqNonNullSingle", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagValuesNonNullLenEq1},
						{Pos: 47, Name: "ElaEqEqNonNullNullSingle", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagNullValuesNonNullLenEq1},
						{Pos: 48, Name: "ElaEqEqNonNull", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagValuesNonNullEq3},
						{Pos: 49, Name: "ElaEqEqNonNullNull", As: MatchedAsDirect, Type: UndTargetTypeElastic, UndTag: tagNullValuesNonNullLenEq3},
					},
				},
			},
			{
				K: 1,
				V: RawMatchedType{
					Pos:     1,
					Name:    "WithTypeParam",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Pos:    2,
							Name:   "Baz",
							As:     MatchedAsDirect,
							Type:   UndTargetTypeOption,
							UndTag: tagRequired,
						},
					},
				},
			},
		},
		types,
	)

	f = pkg.Files[filepath.Join(cwd, filepath.FromSlash("internal/targettypes/ty2.go"))]
	types = hiter.Collect2(iterable.MapSorted[int, RawMatchedType](f.Types).Iter2())

	deepEqualRawMatchedType(
		t,
		[]hiter.KeyValue[int, RawMatchedType]{
			{
				K: 0,
				V: RawMatchedType{
					Pos:     0,
					Name:    "A",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{Name: "A", As: MatchedAsDirect, Type: UndTargetTypeOption},
					},
				},
			},
			{
				K: 1,
				V: RawMatchedType{
					Pos:     1,
					Name:    "B",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Name: "B",
							As:   MatchedAsDirect,
							Type: UndTargetTypeUnd,
						},
					},
				},
			},
			{
				K: 2,
				V: RawMatchedType{
					Pos:     2,
					Name:    "C",
					Variant: MatchedAsSlice,
					Field: []MatchedField{
						{
							As:   MatchedAsDirect,
							Type: UndTargetTypeElastic,
						},
					},
				},
			},
			{
				K: 3,
				V: RawMatchedType{
					Pos:     3,
					Name:    "D",
					Variant: MatchedAsArray,
					Field: []MatchedField{
						{
							As:   MatchedAsDirect,
							Type: UndTargetTypeSliceUnd,
						},
					},
				},
			},
			{
				K: 4,
				V: RawMatchedType{
					Pos:     4,
					Name:    "F",
					Variant: MatchedAsMap,
					Field: []MatchedField{
						{
							As:   MatchedAsDirect,
							Type: UndTargetTypeSliceElastic,
						},
					},
				},
			},
			{
				K: 5,
				V: RawMatchedType{
					Pos:     5,
					Name:    "Parametrized",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Name: "A",
							As:   "direct",
							Type: UndTargetTypeOption,
						},
					},
				},
			},
			{
				K: 7,
				V: RawMatchedType{
					Pos:     7,
					Name:    "IncludesSubTarget",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Name: "Foo",
							As:   MatchedAsImplementor,
							Type: TargetType{
								ImportPath: "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub",
								Name:       "Baz",
							},
						},
					},
				},
			},
			{
				K: 8,
				V: RawMatchedType{
					Pos:     8,
					Name:    "IncludesImplementor",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Name: "Foo",
							As:   MatchedAsImplementor,
							Type: TargetType{
								ImportPath: "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub",
								Name:       "Foo",
							},
						},
					},
				},
			},
			{
				K: 9,
				V: RawMatchedType{
					Pos:     9,
					Name:    "NestedImplementor",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Name: "Foo",
							As:   MatchedAsDirect,
							Type: UndTargetTypeOption,
							Elem: &MatchedField{
								As: MatchedAsImplementor,
								Type: TargetType{
									ImportPath: "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub",
									Name:       "Foo",
								},
							},
						},
					},
				},
			},
			{
				K: 10,
				V: RawMatchedType{
					Pos:     10,
					Name:    "NestedImplementor2",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Name: "Foo",
							As:   MatchedAsImplementor,
							Type: TargetType{
								ImportPath: "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub",
								Name:       "IncludesImplementor",
							},
						},
					},
				},
			},
		},
		types,
	)

	sub := result["github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub"]
	if sub.Pkg.PkgPath != "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub" {
		t.Errorf("wrong path: %s", sub.Pkg.PkgPath)
	}

	f = sub.Files[filepath.Join(cwd, filepath.FromSlash("internal/targettypes/sub/sub.go"))]
	types = hiter.Collect2(iterable.MapSorted[int, RawMatchedType](f.Types).Iter2())

	deepEqualRawMatchedType(
		t,
		[]hiter.KeyValue[int, RawMatchedType]{
			{
				K: 0,
				V: RawMatchedType{
					Pos:     0,
					Name:    "Baz",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Name: "O",
							As:   MatchedAsDirect,
							Type: UndTargetTypeOption,
						},
					},
				},
			},
			{
				K: 1,
				V: RawMatchedType{
					Pos:     1,
					Name:    "IncludesImplementor",
					Variant: MatchedAsStruct,
					Field: []MatchedField{
						{
							Name: "Foo",
							As:   MatchedAsImplementor,
							Type: TargetType{
								ImportPath: "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub2",
								Name:       "Foo",
							},
						},
					},
				},
			},
		},
		types,
	)
}
