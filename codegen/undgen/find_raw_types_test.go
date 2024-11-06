package undgen

import (
	"go/ast"
	"go/types"
	"strings"
	"testing"

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
