package tests

import (
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/ngicks/go-codegen/codegen/generator/cloner"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"gotest.tools/v3/assert"
)

func TestGenerate_nocopy(t *testing.T) {
	pkgs := testTargets["nocopy"]
	testPrinter := suffixwriter.NewTestWriter(".cloner", suffixwriter.WithCwd("../testtargets"))
	cfg := cloner.Config{}
	err := cfg.Generate(
		context.Background(),
		testPrinter.Writer,
		pkgs,
		nil,
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}
