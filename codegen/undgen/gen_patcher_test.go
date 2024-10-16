package undgen

import (
	"maps"
	"slices"
	"testing"

	"github.com/ngicks/go-codegen/codegen/suffixprinter"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

func Test_generatePatcher(t *testing.T) {
	var pkg *packages.Package
	for _, p := range testdataPackages {
		if p.PkgPath == "github.com/ngicks/go-codegen/codegen/undgen/testdata/targettypes" {
			pkg = p
			break
		}
	}

	testPrinter := suffixprinter.NewTestPrinter("yay")
	err := GeneratePatcher(
		testPrinter.Printer,
		pkg,
		ConstUnd.Imports,
		"All", "WithTypeParam", "A", "B", "IncludesSubTarget",
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}
