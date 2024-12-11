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

func TestGenerate_param(t *testing.T) {
	pkgs := testTargets["param"]
	testPrinter := suffixwriter.NewTestWriter(".cloner", suffixwriter.WithCwd("../testtargets"))
	cfg := cloner.Config{
		MatcherConfig: &cloner.MatcherConfig{
			ChannelHandle: cloner.NoCopyHandleDisallow,
		},
	}
	err := cfg.Generate(
		context.Background(),
		testPrinter.Writer,
		pkgs,
	)
	assert.NilError(t, err)
	results := testPrinter.Results()
	for _, k := range slices.Sorted(maps.Keys(results)) {
		result := results[k]
		t.Logf("%q:\n%s", k, result)
	}
}
