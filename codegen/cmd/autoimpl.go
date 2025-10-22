/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/ngicks/go-codegen/codegen/pkg/autoimpl"
	"github.com/spf13/cobra"
)

var (
	// Generator filter flags
	onlyGenerators []string
	skipGenerators []string

	// autoimpl-specific flags
	ignoreGenerated bool
	autoimplDryRun  bool
)

func init() {
	fset := autoimplCmd.Flags()

	// Standard flags
	fset.StringSliceP("pkg", "p", nil, "[required] target package patterns (e.g., ./..., ./pkg/models)")
	_ = autoimplCmd.MarkFlagRequired("pkg")

	fset.StringP("dir", "d", "", "working directory (default: current directory)")
	fset.BoolP("verbose", "v", false, "enable verbose logging")

	// Generator control
	fset.StringSliceVar(&onlyGenerators, "only", nil,
		"only run specific generators (ignore others in markers). Can be specified multiple times.")
	fset.StringSliceVar(&skipGenerators, "skip", nil,
		"skip specific generators even if marked. Can be specified multiple times.")

	// Operation modes
	fset.BoolVar(&autoimplDryRun, "dry-run", false,
		"preview what would be generated without writing files")
	fset.BoolVar(&ignoreGenerated, "ignore-generated", false,
		"skip scanning generated files for markers")

	rootCmd.AddCommand(autoimplCmd)
}

// autoimplCmd represents the autoimpl command
var autoimplCmd = &cobra.Command{
	Use:   "autoimpl [packages...]",
	Short: "Generate implementations for types marked with directive comments",
	Long: `autoimpl scans for types marked with directive comments and generates code.

The autoimpl command implements the second phase of the two-phase workflow:
1. automark: Discover and mark types with directive comments
2. autoimpl: Generate implementations for marked types

This command scans source files for types marked with //codegen: directives
and dispatches to the appropriate generators based on the markers.

Examples:
  # Generate all marked types in package
  codegen autoimpl ./pkg/models

  # Generate all marked types in entire codebase
  codegen autoimpl ./...

  # Preview what would be generated
  codegen autoimpl ./... --dry-run

  # Generate only cloner code (skip others)
  codegen autoimpl ./... --only cloner

  # Skip validator generation
  codegen autoimpl ./... --skip und:validator

  # Ignore generated files when scanning
  codegen autoimpl ./... --ignore-generated

Marker format (added by automark or manually):
  //codegen:cloner no-copy:copy chan:make
  type MyStruct struct {
      Name string
  }

The command reads these markers and generates the appropriate code files:
  - cloner: *.clone.go
  - und:patch: *.und_patch.go
  - und:plain: *.und_plain.go
  - und:validator: *.und_validator.go
`,
	RunE: runAutoimpl,
}

func runAutoimpl(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	fset := cmd.Flags()

	// Get flags
	pkgPatterns, err := fset.GetStringSlice("pkg")
	if err != nil {
		return err
	}

	// If positional args provided, use those instead
	if len(args) > 0 {
		pkgPatterns = args
	}

	dir, err := fset.GetString("dir")
	if err != nil {
		return err
	}
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
	}

	verbose, err := fset.GetBool("verbose")
	if err != nil {
		return err
	}

	// Build configuration
	cfg := &autoimpl.DispatchConfig{
		PackagePatterns:   pkgPatterns,
		GeneratorRegistry: autoimpl.DefaultRegistry(),
		Verbose:           verbose,
		DryRun:            autoimplDryRun,
		IgnoreGenerated:   ignoreGenerated,
		WorkingDir:        dir,
		OnlyGenerators:    onlyGenerators,
		SkipGenerators:    skipGenerators,
	}

	// Validate configuration
	if err := autoimpl.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if verbose {
		fmt.Println("Starting autoimpl...")
		fmt.Printf("Working directory: %s\n", dir)
		fmt.Printf("Package patterns: %v\n", pkgPatterns)
		fmt.Printf("Dry run: %v\n", autoimplDryRun)
		fmt.Printf("Ignore generated: %v\n", ignoreGenerated)
		if len(onlyGenerators) > 0 {
			fmt.Printf("Only generators: %v\n", onlyGenerators)
		}
		if len(skipGenerators) > 0 {
			fmt.Printf("Skip generators: %v\n", skipGenerators)
		}
		fmt.Println()
	}

	// Scan for marked types
	marked, err := autoimpl.ScanForMarkedTypes(ctx, cfg)
	if err != nil {
		return fmt.Errorf("scanning for marked types: %w", err)
	}

	if len(marked) == 0 {
		fmt.Println("No marked types found in specified packages.")
		fmt.Println()
		fmt.Println("Hint: Run 'codegen automark' first to mark types for generation.")
		return nil
	}

	if verbose {
		fmt.Printf("\nFound %d marked type(s)\n", len(marked))
		for i, m := range marked {
			fmt.Printf("  %d: %s (generators: %v)\n", i+1, m.TypeName, m.Generators)
		}
		fmt.Println()
	}

	// Dry run mode
	if autoimplDryRun {
		return executeDryRunAutoimpl(marked, cfg)
	}

	// TODO: Actually invoke generators
	// For now, this is a placeholder
	fmt.Println("Generation not yet fully implemented.")
	fmt.Printf("Would generate code for %d marked type(s)\n", len(marked))

	return nil
}

// executeDryRunAutoimpl executes dry-run mode for autoimpl.
func executeDryRunAutoimpl(marked []autoimpl.MarkedType, cfg *autoimpl.DispatchConfig) error {
	fmt.Println("[DRY RUN] Would generate code for the following types:")
	fmt.Println()

	// Group by generator
	byGenerator := make(map[string][]autoimpl.MarkedType)
	for _, m := range marked {
		for _, gen := range m.Generators {
			// Apply filters
			if len(cfg.OnlyGenerators) > 0 {
				found := false
				for _, only := range cfg.OnlyGenerators {
					if gen == only {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Check skip list
			skip := false
			for _, skipGen := range cfg.SkipGenerators {
				if gen == skipGen {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			byGenerator[gen] = append(byGenerator[gen], m)
		}
	}

	for genType, types := range byGenerator {
		fmt.Printf("%s (%d type(s)):\n", genType, len(types))
		for _, t := range types {
			suffix := getOutputSuffix(genType)
			outputFile := t.FilePath[:len(t.FilePath)-3] + suffix // Replace .go with suffix
			fmt.Printf("  %s (%s:%d) → %s\n", t.TypeName, t.FilePath, t.Line, outputFile)
		}
		fmt.Println()
	}

	fmt.Printf("Summary:\n")
	totalTypes := 0
	for _, types := range byGenerator {
		totalTypes += len(types)
	}
	fmt.Printf("  Total types: %d\n", totalTypes)
	fmt.Printf("  Generators: %d\n", len(byGenerator))
	fmt.Printf("\nNo files modified (dry run mode)\n")

	return nil
}

// getOutputSuffix returns the output file suffix for a generator.
func getOutputSuffix(genType string) string {
	switch genType {
	case "cloner":
		return ".clone.go"
	case "und:patch":
		return ".und_patch.go"
	case "und:plain":
		return ".und_plain.go"
	case "und:validator":
		return ".und_validator.go"
	default:
		return ".generated.go"
	}
}
