/*
Copyright © 2024 ngicks <yknt.bsl@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/ngicks/go-codegen/codegen/pkg/automark"
	"github.com/spf13/cobra"
)

var (
	// Generator selection flags
	generators []string

	// Generator configuration flags for cloner
	noCopyConfig    string
	chanConfig      string
	funcConfig      string
	interfaceConfig string

	// Type filtering flags
	includePatterns []string
	excludePatterns []string
	exportedOnly    bool

	// Operation mode flags
	dryRun bool
	force  bool
)

func init() {
	fset := automarkCmd.Flags()

	// Standard flags (reuse existing common flags pattern)
	fset.StringSliceP("pkg", "p", nil, "[required] target package patterns (e.g., ./..., ./pkg/models)")
	_ = automarkCmd.MarkFlagRequired("pkg")

	fset.StringP("dir", "d", "", "working directory (default: current directory)")
	fset.BoolP("verbose", "v", false, "enable verbose logging")

	// Generator selection
	fset.StringSliceVarP(&generators, "generator", "g", nil,
		"[required] generator type to mark types for (cloner, und:patch, und:plain, und:validator). Can be specified multiple times.")
	_ = automarkCmd.MarkFlagRequired("generator")

	// Generator configuration (cloner-specific)
	fset.StringVar(&noCopyConfig, "no-copy", "",
		"how to handle no-copy types (ignore, disallow, copy)")
	fset.StringVar(&chanConfig, "chan", "",
		"how to handle channel fields (ignore, disallow, copy, make)")
	fset.StringVar(&funcConfig, "func", "",
		"how to handle function fields (ignore, disallow, copy)")
	fset.StringVar(&interfaceConfig, "interface", "",
		"how to handle interface fields (ignore, copy)")

	// Type filtering
	fset.StringSliceVar(&includePatterns, "include", nil,
		"only mark types matching these glob patterns (can be specified multiple times)")
	fset.StringSliceVar(&excludePatterns, "exclude", nil,
		"exclude types matching these glob patterns (can be specified multiple times)")
	fset.BoolVar(&exportedOnly, "exported-only", false,
		"only mark exported (capitalized) types")

	// Operation modes
	fset.BoolVar(&dryRun, "dry-run", false,
		"preview which types would be marked without modifying files")
	fset.BoolVar(&force, "force", false,
		"overwrite existing markers even if already present")

	rootCmd.AddCommand(automarkCmd)
}

// automarkCmd represents the automark command
var automarkCmd = &cobra.Command{
	Use:   "automark [packages...]",
	Short: "Discover and mark types with directive comments for code generation",
	Long: `automark discovers types and marks them with directive comments for code generation.

The automark command implements the first phase of the two-phase workflow:
1. automark: Discover and mark types with directive comments
2. autoimpl: Generate implementations for marked types

This separation allows you to review which types will be generated before committing
and mix automatic marking with manual directive editing.

Examples:
  # Mark all types in package for cloner generation
  codegen automark ./pkg/models -g cloner

  # Mark types with specific configuration
  codegen automark ./pkg/services -g cloner --no-copy copy --chan make

  # Mark for multiple generators
  codegen automark ./pkg/types -g und:patch -g und:plain -g und:validator

  # Preview changes without modifying files
  codegen automark ./... -g cloner --dry-run

  # Selective marking with filters
  codegen automark ./... -g cloner --include '*Request' --include '*Response' --exported-only

  # Force re-marking with new configuration
  codegen automark ./pkg/models -g cloner --no-copy copy --force

Markers are added as comments above type declarations:
  //codegen:cloner no-copy:copy chan:make
  type MyStruct struct {
      Name string
  }

After marking, run 'codegen autoimpl' to generate the actual code.
`,
	RunE: runAutomark,
}

func runAutomark(cmd *cobra.Command, args []string) error {
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

	// Build generator specs
	generatorSpecs, err := buildGeneratorSpecs()
	if err != nil {
		return err
	}

	// Build configuration
	cfg := &automark.MarkerConfig{
		PackagePatterns: pkgPatterns,
		Generators:      generatorSpecs,
		TypeFilters: automark.TypeFilterConfig{
			IncludeTypes:    includePatterns,
			ExcludeTypes:    excludePatterns,
			RequireExported: exportedOnly,
		},
		DryRun:     dryRun,
		Verbose:    verbose,
		WorkingDir: dir,
		Force:      force,
	}

	// Execute automark
	result, err := automark.Mark(ctx, cfg)
	if err != nil {
		return fmt.Errorf("automark failed: %w", err)
	}

	// Print summary
	if !dryRun {
		fmt.Println()
		fmt.Println("Automark Summary:")
		fmt.Printf("  Types discovered: %d\n", result.TotalDiscovered)
		fmt.Printf("  After filtering: %d\n", result.TotalFiltered)
		fmt.Printf("  Newly marked: %d\n", result.TotalMarked)
		fmt.Printf("  Skipped (already marked): %d\n", result.TotalSkipped)
		fmt.Printf("  Files modified: %d\n", len(result.FilesModified))

		if result.HasErrors() {
			fmt.Printf("  Errors: %d\n", len(result.Errors))
			fmt.Println("\nErrors:")
			for i, err := range result.Errors {
				fmt.Printf("  %d. %v\n", i+1, err)
			}
			return fmt.Errorf("automark completed with errors")
		}
	}

	return nil
}

// buildGeneratorSpecs constructs GeneratorSpec slice from flags.
func buildGeneratorSpecs() ([]automark.GeneratorSpec, error) {
	if len(generators) == 0 {
		return nil, fmt.Errorf("no generators specified (use --generator/-g flag)")
	}

	var specs []automark.GeneratorSpec

	for _, genType := range generators {
		spec := automark.GeneratorSpec{
			Type:   genType,
			Config: make(map[string]string),
		}

		// Add generator-specific configuration
		if genType == automark.GeneratorCloner {
			if noCopyConfig != "" {
				spec.Config["no-copy"] = noCopyConfig
			}
			if chanConfig != "" {
				spec.Config["chan"] = chanConfig
			}
			if funcConfig != "" {
				spec.Config["func"] = funcConfig
			}
			if interfaceConfig != "" {
				spec.Config["interface"] = interfaceConfig
			}
		}

		specs = append(specs, spec)
	}

	return specs, nil
}
