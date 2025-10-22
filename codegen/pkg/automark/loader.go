package automark

import (
	"context"
	"fmt"

	"github.com/ngicks/go-codegen/codegen/pkg/astutil"
	"github.com/ngicks/go-codegen/codegen/pkg/pkgsutil"
	"golang.org/x/tools/go/packages"
)

// LoadPackages loads Go packages based on the configuration.
func LoadPackages(ctx context.Context, cfg *MarkerConfig) ([]*packages.Package, error) {
	if len(cfg.PackagePatterns) == 0 {
		return nil, fmt.Errorf("no package patterns specified")
	}

	pkgCfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
		Context: ctx,
		Dir:     cfg.WorkingDir,
	}

	if cfg.Verbose {
		pkgCfg.Logf = func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		}
	}

	// Use custom parser to exclude generated files if needed
	// (though for marking, we typically want to see all files)
	pkgCfg.ParseFile = astutil.NewParser(cfg.WorkingDir).ParseFile

	pkgs, err := packages.Load(pkgCfg, cfg.PackagePatterns...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	if err := pkgsutil.CheckLoadError(pkgs); err != nil {
		return nil, fmt.Errorf("package load errors: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found matching patterns: %v", cfg.PackagePatterns)
	}

	if cfg.Verbose {
		fmt.Printf("\nLoaded %d package(s):\n", len(pkgs))
		for i, pkg := range pkgs {
			fmt.Printf("  %d: %s\n", i+1, pkg.PkgPath)
		}
		fmt.Println()
	}

	return pkgs, nil
}
