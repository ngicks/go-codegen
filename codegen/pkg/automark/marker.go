package automark

import (
	"context"
	"fmt"
	"go/types"
)

// TypeMarker represents a directive comment that marks a type for code generation.
// It is the output of automark and input to autoimpl.
type TypeMarker struct {
	// TypeName is the name of the marked type
	TypeName string

	// Type is the Go type object
	Type types.Type

	// Generators is the list of generators to apply
	Generators []GeneratorSpec

	// Location contains file and position information
	Location SourceLocation

	// ExistingDirectives contains any existing field-level directives
	ExistingDirectives []FieldDirective
}

// SourceLocation tracks where in source code a marker exists.
type SourceLocation struct {
	// FilePath is the absolute path to the file
	FilePath string

	// Package is the Go package path
	Package string

	// Line is the line number of the type declaration
	Line int

	// Column is the column number
	Column int
}

// FieldDirective represents an existing per-field customization directive.
// These are preserved by automark and used by autoimpl.
type FieldDirective struct {
	// Field is the field name
	Field string

	// Generator is the generator name (e.g., "cloner", "undgen")
	Generator string

	// Directive is the directive name (e.g., "copyptr", "make", "ignore")
	Directive string

	// Config contains additional key=value configuration
	Config map[string]string
}

// Mark is the main entry point for the automark functionality.
// It discovers types, applies filters, and writes markers to source files.
func Mark(ctx context.Context, cfg *MarkerConfig) (*MarkResult, error) {
	// Validate configuration
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if cfg.Verbose {
		fmt.Println("Starting automark...")
		fmt.Printf("Working directory: %s\n", cfg.WorkingDir)
		fmt.Printf("Package patterns: %v\n", cfg.PackagePatterns)
		fmt.Printf("Generators: %d\n", len(cfg.Generators))
		fmt.Printf("Filters: %s\n", FormatFilterSummary(cfg.TypeFilters))
		fmt.Printf("Dry run: %v\n", cfg.DryRun)
		fmt.Printf("Force: %v\n", cfg.Force)
		fmt.Println()
	}

	// Load packages
	pkgs, err := LoadPackages(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	// Discover types
	candidates, err := DiscoverTypes(pkgs, cfg)
	if err != nil {
		return nil, fmt.Errorf("discovering types: %w", err)
	}

	if cfg.Verbose {
		fmt.Printf("Found %d type candidates\n", len(candidates))
	}

	// Apply filters
	filtered := ApplyFilters(candidates, cfg.TypeFilters)

	if cfg.Verbose {
		fmt.Printf("After filtering: %d types eligible for marking\n", len(filtered))
		if len(filtered) != len(candidates) {
			fmt.Printf("Filtered out: %d types\n", len(candidates)-len(filtered))
		}
		fmt.Println()
	}

	// Create marker plans for each generator
	var plans []TypeMarkerPlan
	for _, candidate := range filtered {
		for _, gen := range cfg.Generators {
			// Check if already marked (unless force mode)
			if !cfg.Force && HasExistingMarker(candidate.GenDecl, gen.Type) {
				if cfg.Verbose {
					fmt.Printf("Skipping %s for %s (already marked)\n", candidate.TypeName, gen.Type)
				}
				continue
			}

			plans = append(plans, TypeMarkerPlan{
				Candidate: candidate,
				Spec:      gen,
			})
		}
	}

	if len(plans) == 0 {
		return &MarkResult{
			TotalDiscovered: len(candidates),
			TotalFiltered:   len(filtered),
			TotalMarked:     0,
			TotalSkipped:    len(filtered),
		}, nil
	}

	if cfg.Verbose {
		fmt.Printf("Planning to mark %d type-generator combination(s)\n\n", len(plans))
	}

	// Dry run mode - just report what would be done
	if cfg.DryRun {
		return executeDryRun(plans, cfg), nil
	}

	// Write markers to files
	writeResult, err := WriteMarkers(pkgs, plans, cfg)
	if err != nil {
		return nil, fmt.Errorf("writing markers: %w", err)
	}

	result := &MarkResult{
		TotalDiscovered: len(candidates),
		TotalFiltered:   len(filtered),
		TotalMarked:     writeResult.TotalMarked,
		TotalSkipped:    len(plans) - writeResult.TotalMarked,
		FilesModified:   writeResult.FilesModified,
		Errors:          writeResult.Errors,
	}

	return result, nil
}

// executeDryRun executes dry-run mode and returns results without modifying files.
func executeDryRun(plans []TypeMarkerPlan, cfg *MarkerConfig) *MarkResult {
	fmt.Println("[DRY RUN] Would mark the following types:")
	fmt.Println()

	// Group by file for better output
	byFile := make(map[string][]TypeMarkerPlan)
	for _, plan := range plans {
		path := plan.Candidate.Location.FilePath
		byFile[path] = append(byFile[path], plan)
	}

	for filePath, filePlans := range byFile {
		fmt.Printf("%s:\n", filePath)
		for _, plan := range filePlans {
			marker := FormatMarker(plan.Spec)
			fmt.Printf("  %s (line %d) → //%s\n",
				plan.Candidate.TypeName,
				plan.Candidate.Location.Line,
				marker,
			)
		}
		fmt.Println()
	}

	fmt.Printf("Summary:\n")
	fmt.Printf("  Total types to mark: %d\n", len(plans))
	fmt.Printf("  Files to modify: %d\n", len(byFile))
	fmt.Printf("\nNo files modified (dry run mode)\n")

	return &MarkResult{
		TotalDiscovered: 0, // Not tracked in dry run
		TotalFiltered:   0, // Not tracked in dry run
		TotalMarked:     len(plans),
		TotalSkipped:    0,
		FilesModified:   nil,
		Errors:          nil,
	}
}

// MarkResult contains the results of a Mark operation.
type MarkResult struct {
	TotalDiscovered int
	TotalFiltered   int
	TotalMarked     int
	TotalSkipped    int
	FilesModified   map[string]int
	Errors          []error
}

// HasErrors returns true if there were any errors.
func (r *MarkResult) HasErrors() bool {
	return len(r.Errors) > 0
}
