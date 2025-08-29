package tests

import (
	"maps"
	"slices"
	"testing"

	"github.com/ngicks/go-codegen/codegen/pkg/suffixwriter"
	"github.com/ngicks/go-codegen/codegen/generator/undgen"
	"gotest.tools/v3/assert"
)

func Test_display_patcher(t *testing.T) {
	pkgs := testTargets["display"]
	testPrinter := suffixwriter.NewTestWriter(".und_patcher", suffixwriter.WithCwd("../testtargets"))
	err := undgen.GeneratePatcher(
		testPrinter.Writer,
		true,
		pkgs[0],
		undgen.ConstUnd.Imports,
		"...",
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}

func Test_display_validator(t *testing.T) {
	pkgs := testTargets["display"]
	testPrinter := suffixwriter.NewTestWriter(".und_validator", suffixwriter.WithCwd("../testtargets"))
	err := undgen.GenerateValidator(
		testPrinter.Writer,
		true,
		pkgs,
		undgen.ConstUnd.Imports,
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}

func Test_display_plain(t *testing.T) {
	pkgs := testTargets["display"]
	testPrinter := suffixwriter.NewTestWriter(".und_plain", suffixwriter.WithCwd("../testtargets"))
	err := undgen.GeneratePlain(
		testPrinter.Writer,
		true,
		pkgs,
		undgen.ConstUnd.Imports,
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}
