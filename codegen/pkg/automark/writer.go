package automark

import (
	"fmt"
	"go/format"
	"go/token"
	"os"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/packages"
)

// WriteMarkers writes type markers to source files using DST to preserve formatting.
func WriteMarkers(pkgs []*packages.Package, markers []TypeMarkerPlan, cfg *MarkerConfig) (WriteResult, error) {
	result := WriteResult{
		FilesModified: make(map[string]int),
	}

	// Group markers by file
	byFile := make(map[string][]TypeMarkerPlan)
	for _, marker := range markers {
		byFile[marker.Candidate.Location.FilePath] = append(byFile[marker.Candidate.Location.FilePath], marker)
	}

	if cfg.Verbose {
		fmt.Printf("\nProcessing %d file(s) with markers\n", len(byFile))
	}

	for filePath, fileMarkers := range byFile {
		if cfg.Verbose {
			fmt.Printf("  Processing %s (%d markers)\n", filePath, len(fileMarkers))
		}

		count, err := writeMarkersToFile(filePath, fileMarkers, cfg)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", filePath, err))
			continue
		}

		result.FilesModified[filePath] = count
		result.TotalMarked += count
	}

	return result, nil
}

// writeMarkersToFile writes markers to a single file using DST.
func writeMarkersToFile(filePath string, markers []TypeMarkerPlan, cfg *MarkerConfig) (int, error) {
	// Read the file
	src, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("reading file: %w", err)
	}

	// Parse with DST to preserve formatting
	dec := decorator.NewDecorator(nil)
	file, err := dec.Parse(src)
	if err != nil {
		return 0, fmt.Errorf("parsing file: %w", err)
	}

	count := 0

	// Add markers to the file
	for _, marker := range markers {
		// Skip if should not mark (already exists and not force mode)
		if ShouldSkipMarking(marker.Candidate.GenDecl, marker.Spec.Type, cfg.Force) {
			if cfg.Verbose {
				fmt.Printf("    Skipping %s (already marked)\n", marker.Candidate.TypeName)
			}
			continue
		}

		// Format the marker comment
		markerText := FormatMarker(marker.Spec)

		// Add the marker using DST
		if err := addMarkerComment(file, marker.Candidate, markerText, cfg); err != nil {
			return count, fmt.Errorf("adding marker to %s: %w", marker.Candidate.TypeName, err)
		}

		count++
		if cfg.Verbose {
			fmt.Printf("    Marked %s with %s\n", marker.Candidate.TypeName, markerText)
		}
	}

	// If dry-run, don't write the file
	if cfg.DryRun {
		return count, nil
	}

	// Write the file back
	if count > 0 {
		res := decorator.NewRestorer()
		astFile, err := res.RestoreFile(file)
		if err != nil {
			return count, fmt.Errorf("restoring AST: %w", err)
		}

		// Open file for writing
		f, err := os.Create(filePath)
		if err != nil {
			return count, fmt.Errorf("creating file: %w", err)
		}
		defer f.Close()

		// Print AST to file
		fset := token.NewFileSet()
		if err := format.Node(f, fset, astFile); err != nil {
			return count, fmt.Errorf("formatting file: %w", err)
		}
	}

	return count, nil
}

// addMarkerComment adds a marker comment to a type declaration using DST.
func addMarkerComment(file *dst.File, candidate TypeCandidate, markerText string, cfg *MarkerConfig) error {
	// Find the GenDecl in the DST file that corresponds to this candidate
	// We need to match by position

	var found bool
	dst.Inspect(file, func(n dst.Node) bool {
		if found {
			return false
		}

		genDecl, ok := n.(*dst.GenDecl)
		if !ok {
			return true
		}

		// Check if this is the right declaration by checking type specs
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*dst.TypeSpec)
			if !ok {
				continue
			}

			if typeSpec.Name.Name == candidate.TypeName {
				// Found it! Add the marker comment
				if genDecl.Decs.Start == nil {
					genDecl.Decs.Start = dst.Decorations{}
				}

				// Check if in force mode and need to remove old marker
				if cfg.Force {
					// Remove any existing codegen: comments
					var filtered []string
					for _, line := range genDecl.Decs.Start {
						trimmed := strings.TrimSpace(line)
						if !strings.HasPrefix(strings.TrimPrefix(trimmed, "//"), TypeLevelPrefix) {
							filtered = append(filtered, line)
						}
					}
					genDecl.Decs.Start = filtered
				}

				// Add new marker at the beginning
				genDecl.Decs.Start.Prepend("//" + markerText)

				found = true
				return false
			}
		}

		return true
	})

	if !found {
		return fmt.Errorf("type declaration not found in DST")
	}

	return nil
}

// FormatMarker formats a GeneratorSpec into a marker comment string.
// Returns just the content after "//", e.g., "codegen:cloner no-copy:copy"
func FormatMarker(spec GeneratorSpec) string {
	parts := []string{TypeLevelPrefix + spec.Type}

	// Add config options in sorted order for consistency
	var keys []string
	for k := range spec.Config {
		keys = append(keys, k)
	}

	// Sort keys for deterministic output
	// Using simple sort for now
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, key := range keys {
		value := spec.Config[key]
		if value != "" {
			parts = append(parts, fmt.Sprintf("%s:%s", key, value))
		} else {
			parts = append(parts, key)
		}
	}

	return strings.Join(parts, " ")
}

// TypeMarkerPlan represents a plan to mark a type.
type TypeMarkerPlan struct {
	Candidate TypeCandidate
	Spec      GeneratorSpec
}

// WriteResult contains the results of a write operation.
type WriteResult struct {
	FilesModified map[string]int
	TotalMarked   int
	Errors        []error
}

// HasErrors returns true if there were any errors during writing.
func (r WriteResult) HasErrors() bool {
	return len(r.Errors) > 0
}
