package loader

import (
	"fmt"
	"testing"

	"golang.org/x/tools/go/packages"
)

var cfg = &packages.Config{
	Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes |
		packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedDeps |
		packages.NeedImports,
}

func LoadPackagesPanicking(patterns ...string) []*packages.Package {
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			panic(fmt.Errorf("package %s has error: %v", pkg.PkgPath, err))
		}
	}
	return pkgs
}

func LoadPackagesTest(t *testing.T, patterns ...string) []*packages.Package {
	t.Helper()
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		t.Fatalf("failed to load packages: %v", err)
	}
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			t.Fatalf("package %s has error: %v", pkg.PkgPath, err)
		}
	}
	return pkgs
}
