package undgen

import (
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"slices"
	"testing"

	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/packages"
	"gotest.tools/v3/assert"
)

func Test_generatePatcher(t *testing.T) {
	var pkg *packages.Package
	for _, p := range testdataPackages {
		if p.PkgPath == "github.com/ngicks/und/internal/undgen/testdata/targettypes" {
			pkg = p
			break
		}
	}

	for data, err := range generatePatcher(
		pkg,
		ConstUnd.Imports,
		"WithTypeParam", "A", "B", "IncludesSubTarget",
	) {
		assert.NilError(t, err)
		res := decorator.NewRestorer()
		afile, err := res.RestoreFile(data.df)
		assert.NilError(t, err)

		for _, dec := range afile.Decls {
			genDecl, ok := dec.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				ts := spec.(*ast.TypeSpec)
				if !slices.Contains(data.typeNames, ts.Name.Name) {
					continue
				}
				os.Stdout.Write([]byte("\n"))
				printer.Fprint(os.Stdout, res.Fset, ts)
				os.Stdout.Write([]byte("\n"))
			}
		}
	}

}
