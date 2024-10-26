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

func Test_generateValidator1(t *testing.T) {
	testPrinter := suffixwriter.NewTestWriter(".und_validator")
	err := GenerateValidator(
		testPrinter.Writer,
		true,
		slices.Collect(
			xiter.Filter(func(pkg *packages.Package) bool {
				return pkg.PkgPath != "github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/erroneous"
			},
				slices.Values(targettypesPackages),
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

func Test_generateValidator2(t *testing.T) {
	testPrinter := suffixwriter.NewTestWriter(".und_validator")
	err := GenerateValidator(
		testPrinter.Writer,
		true,
		validatorPackages,
		ConstUnd.Imports,
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}
