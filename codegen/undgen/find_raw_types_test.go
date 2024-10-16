package undgen

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/hiter/iterable"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

var (
	testdataPackages []*packages.Package
)

func init() {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedExportFile |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedModule |
			packages.NeedEmbedFiles |
			packages.NeedEmbedPatterns,
		Logf: func(format string, args ...interface{}) {
			fmt.Printf("log: "+format, args...)
			fmt.Println()
		},
	}
	var err error
	testdataPackages, err = packages.Load(cfg, "./testdata/targettypes/...")
	if err != nil {
		panic(err)
	}
}

func Test_isImplementor(t *testing.T) {
	var foo, fooPlain, bar, nonCyclic *types.Named

	for _, pkg := range testdataPackages {
		if pkg.PkgPath != "github.com/ngicks/und/internal/undgen/testdata/targettypes/sub" {
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
	assert.Assert(t, isImplementor(foo, mset, false))
	assert.Assert(t, isImplementor(fooPlain, mset, true))

	assert.Assert(t, !isImplementor(bar, mset, true))
	assert.Assert(t, !isImplementor(nonCyclic, mset, true))
}

func Test_parseImports(t *testing.T) {
	var file1, file2 *ast.File
P:
	for _, p := range testdataPackages {
		for _, f := range p.Syntax {
			fPath := p.Fset.Position(f.FileStart)
			if strings.HasSuffix(fPath.Filename, "internal/undgen/testdata/targettypes/ty1.go") {
				file1 = f
			}
			if strings.HasSuffix(fPath.Filename, "internal/undgen/testdata/targettypes/ty2.go") {
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
		fallback: map[string]TargetImport{
			"elastic_1": ConstUnd.Imports[4],
		},
	}
	assert.DeepEqual(
		t,
		expected.identToImport,
		importMap.identToImport,
	)
	assert.DeepEqual(
		t,
		expected.fallback,
		importMap.fallback,
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
		fallback: map[string]TargetImport{},
	}
	assert.DeepEqual(
		t,
		expected.identToImport,
		importMap.identToImport,
	)
	assert.DeepEqual(
		t,
		expected.fallback,
		importMap.fallback,
	)
}

func TestFindTargetType_error(t *testing.T) {
	result, err := FindRawTypes(
		testdataPackages,
		ConstUnd.Imports,
		ConstUnd.ConversionMethod,
	)
	assert.Assert(t, err != nil)
	assert.Assert(t, len(result) > 0)
}

func deepEqualRawMatchedType(t *testing.T, i, j []hiter.KeyValue[int, RawMatchedType]) {
	assert.DeepEqual(
		t,
		i, j,
		// generally ignore additional fields.
		// other code / tests use these so isn't too bad to ignore here.
		gocmp.Comparer(func(i, j *ast.GenDecl) bool { return true }),
		gocmp.Comparer(func(i, j *ast.TypeSpec) bool { return true }),
		gocmp.Comparer(func(i, j types.Object) bool { return true }),
	)
}

func TestFindTargetType(t *testing.T) {
	result, err := FindRawTypes(
		slices.Collect(
			xiter.Filter(func(pkg *packages.Package) bool {
				return pkg.PkgPath != "github.com/ngicks/und/internal/undgen/testdata/targettypes/erroneous"
			},
				slices.Values(testdataPackages),
			),
		),
		ConstUnd.Imports,
		ConstUnd.ConversionMethod,
	)
	assert.NilError(t, err)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pkg := result["github.com/ngicks/und/internal/undgen/testdata/targettypes"]
	if pkg.Pkg.PkgPath != "github.com/ngicks/und/internal/undgen/testdata/targettypes" {
		t.Errorf("wrong path: %s", pkg.Pkg.PkgPath)
	}

	f := pkg.Files[filepath.Join(cwd, filepath.FromSlash("testdata/targettypes/ty1.go"))]
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
						{Name: "UntouchedOpt", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "UntouchedUnd", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UntouchedSliceUnd", As: MatchedAsDirect, Type: UndTargetTypesSliceUnd},
						{Name: "OptRequired", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "OptNullish", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "OptDef", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "OptNull", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "OptUnd", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "OptDefOrUnd", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "OptDefOrNull", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "OptNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "OptDefOrNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeOption},
						{Name: "UndRequired", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UndNullish", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UndDef", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UndNull", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UndUnd", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UndDefOrUnd", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UndDefOrNull", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UndNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "UndDefOrNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypesUnd},
						{Name: "ElaRequired", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaNullish", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaDef", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaNull", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaDefOrUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaDefOrNull", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaDefOrNullOrUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEq", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaGr", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaGrEq", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaLe", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaLeEq", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEquRequired", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEquNullish", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEquDef", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEquNull", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEquUnd", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEqNonNullSlice", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEqNonNullNullSlice", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEqNonNullSingle", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEqNonNullNullSingle", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEqNonNull", As: MatchedAsDirect, Type: UndTargetTypeElastic},
						{Name: "ElaEqEqNonNullNull", As: MatchedAsDirect, Type: UndTargetTypeElastic},
					},
				},
			},
			{
				K: 1,
				V: RawMatchedType{
					Pos:     1,
					Name:    "WithTypeParam",
					Variant: "struct",
					Field: []MatchedField{
						{
							Name: "Baz",
							As:   MatchedAsDirect,
							Type: UndTargetTypeOption,
						},
					},
				},
			},
		},
		types,
	)

	f = pkg.Files[filepath.Join(cwd, filepath.FromSlash("testdata/targettypes/ty2.go"))]
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
							Type: UndTargetTypesUnd,
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
							Type: UndTargetTypesSliceUnd,
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
							Type: UndTargetTypesSliceElastic,
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
							Elem: MatchedFieldElem{
								As: MatchedAsImplementor,
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
						},
					},
				},
			},
		},
		types,
	)

	sub := result["github.com/ngicks/und/internal/undgen/testdata/targettypes/sub"]
	if sub.Pkg.PkgPath != "github.com/ngicks/und/internal/undgen/testdata/targettypes/sub" {
		t.Errorf("wrong path: %s", sub.Pkg.PkgPath)
	}

	f = sub.Files[filepath.Join(cwd, filepath.FromSlash("testdata/targettypes/sub/sub.go"))]
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
					Variant: "struct",
					Field: []MatchedField{
						{
							Name: "Foo",
							As:   MatchedAsImplementor,
						},
					},
				},
			},
		},
		types,
	)
}
