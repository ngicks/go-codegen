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

	var names []string
	for _, dirent := range dirents {
		name := dirent.Name()
		if !dirent.IsDir() || slices.Contains(strings.Split(*excludes, ","), name) {
			continue
		}
		names = append(names, name)
		f, err := os.Create(name + "_generate_test.go")
		if err != nil {
			panic(err)
		}
		_, err = fmt.Fprintf(
			f,
			`package tests

import (
	"maps"
	"slices"
	"testing"

	"github.com/ngicks/go-codegen/codegen/suffixwriter"
	"github.com/ngicks/go-codegen/codegen/generator/undgen"
	"gotest.tools/v3/assert"
)

func Test_%[1]s_patcher(t *testing.T) {
	pkgs := testTargets["%[1]s"]
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
		t.Logf("%%q:\n%%s", k, result)
	}
}

func Test_%[1]s_validator(t *testing.T) {
	pkgs := testTargets["%[1]s"]
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
		t.Logf("%%q:\n%%s", k, result)
	}
}

func Test_%[1]s_plain(t *testing.T) {
	pkgs := testTargets["%[1]s"]
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
