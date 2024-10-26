package undgen

import (
	"maps"
	"slices"
	"testing"

	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

func TestGenPlain(t *testing.T) {
	testPrinter := suffixwriter.NewTestWriter(".und_plain")
	err := GeneratePlain(
		testPrinter.Writer,
		true,
		slices.Collect(
			xiter.Filter(func(pkg *packages.Package) bool {
				return pkg.PkgPath != "github.com/ngicks/go-codegen/codegen/undgen/testdata/targettypes/erroneous"
			},
				slices.Values(testdataPackages),
			),
		),
		ConstUnd.Imports,
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}