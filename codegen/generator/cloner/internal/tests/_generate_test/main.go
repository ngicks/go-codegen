package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
)

var (
	excludes = flag.String("e", "", "")
)

func main() {
	flag.Parse()

	dirents, err := os.ReadDir("../testtargets")
	if err != nil {
		panic(err)
	}

	for _, dirent := range dirents {
		name := dirent.Name()
		if !dirent.IsDir() || slices.Contains(strings.Split(*excludes, ","), name) {
			continue
		}
		f, err := os.Create(name + "_generate_test.go")
		if err != nil {
			panic(err)
		}
		_, err = fmt.Fprintf(
			f,
			`package tests

import (
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/ngicks/go-codegen/codegen/generator/cloner"
	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"gotest.tools/v3/assert"
)

func TestGenerate_%[1]s(t *testing.T) {
	pkgs := testTargets[%[1]q]
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
		t.Logf("%%q:\n%%s", k, result)
	}
}
`,
			name,
		)
		if err != nil {
			panic(err)
		}
	}
}
