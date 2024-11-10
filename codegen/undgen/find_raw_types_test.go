package undgen

import (
	"go/types"
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
