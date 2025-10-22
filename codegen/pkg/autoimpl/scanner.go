package autoimpl

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/ngicks/go-codegen/codegen/pkg/astutil"
	"github.com/ngicks/go-codegen/codegen/pkg/directive"
	"github.com/ngicks/go-codegen/codegen/pkg/pkgsutil"
	"golang.org/x/tools/go/packages"
)

// ScanForMarkedTypes scans packages for types marked with codegen directives.
func ScanForMarkedTypes(ctx context.Context, cfg *DispatchConfig) ([]MarkedType, error) {
	// Load packages
	pkgs, err := loadPackages(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	var marked []MarkedType

	for _, pkg := range pkgs {
		pkgMarked, err := scanPackage(pkg, cfg)
		if err != nil {
			return nil, fmt.Errorf("scanning package %s: %w", pkg.PkgPath, err)
		}
		marked = append(marked, pkgMarked...)
	}

	if cfg.Verbose {
		fmt.Printf("Found %d marked type(s)\n", len(marked))
	}

	return marked, nil
}

// loadPackages loads packages based on configuration.
func loadPackages(ctx context.Context, cfg *DispatchConfig) ([]*packages.Package, error) {
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

	// Optionally ignore generated files
	if cfg.IgnoreGenerated {
		pkgCfg.ParseFile = astutil.NewParser(cfg.WorkingDir).ParseFile
	}

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

// scanPackage scans a single package for marked types.
func scanPackage(pkg *packages.Package, cfg *DispatchConfig) ([]MarkedType, error) {
	var marked []MarkedType

	for _, file := range pkg.Syntax {
		filePath := pkg.Fset.File(file.Pos()).Name()

		// Inspect AST for type declarations with markers
		ast.Inspect(file, func(n ast.Node) bool {
			genDecl, ok := n.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				return true
			}

			// Check if this declaration should be included
			// Note: ExcludeIgnoredGenDecl returns true to INCLUDE, false to EXCLUDE
			include, err := directive.ExcludeIgnoredGenDecl(genDecl)
			if err != nil || !include {
				return true
			}

			// Check for codegen markers in comments
			markers := parseTypeMarkers(genDecl.Doc)
			if len(markers) == 0 {
				return true
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// Get type object for ExcludeIgnoredTypeSpec
				obj := pkg.TypesInfo.Defs[typeSpec.Name]
				if obj == nil {
					continue
				}

				// Check if this type spec should be ignored
				exclude, err := directive.ExcludeIgnoredTypeSpec(typeSpec, obj)
				if err != nil || exclude {
					continue
				}

				typeName := typeSpec.Name.Name
				pos := pkg.Fset.Position(typeSpec.Pos())

				// Create MarkedType for each generator in markers
				for _, marker := range markers {
					markedType := MarkedType{
						TypeName:   typeName,
						Package:    pkg,
						Generators: []string{marker.GeneratorType},
						Config:     marker.Config,
						FilePath:   filePath,
						Line:       pos.Line,
					}

					marked = append(marked, markedType)

					if cfg.Verbose {
						fmt.Printf("Found marked type: %s (generator: %s) at %s:%d\n",
							typeName, marker.GeneratorType, filePath, pos.Line)
					}
				}
			}

			return true
		})
	}

	return marked, nil
}
