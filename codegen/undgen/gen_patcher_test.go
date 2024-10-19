package undgen

import (
	"log/slog"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/ngicks/go-codegen/codegen/suffixwriter"
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

	testPrinter := suffixwriter.NewTestWriter("yay")
	err := GeneratePatcher(
		testPrinter.Writer,
		true,
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

func Test_generatePatcher_write(t *testing.T) {
	var pkg *packages.Package
	for _, p := range patchtargetPackages {
		if p.PkgPath == "github.com/ngicks/go-codegen/codegen/undgen/testdata/patchtarget" {
			pkg = p
			break
		}
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	writer := suffixwriter.New(".und_patch")
	err := GeneratePatcher(
		writer,
		true,
		pkg,
		ConstUnd.Imports,
		"All", "Ignored", "Hmm", "NameOverlapping",
	)
	assert.NilError(t, err)
}
