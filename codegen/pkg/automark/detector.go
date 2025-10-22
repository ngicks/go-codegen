package automark

import (
	"go/ast"
	"strings"
)

// HasExistingMarker checks if a type already has a marker for the specified generator.
func HasExistingMarker(genDecl *ast.GenDecl, generatorType string) bool {
	if genDecl.Doc == nil {
		return false
	}

	markerPrefix := TypeLevelPrefix + generatorType

	for _, comment := range genDecl.Doc.List {
		text := strings.TrimSpace(comment.Text)
		// Remove comment markers
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		// Check if this is our marker
		if strings.HasPrefix(text, markerPrefix) {
			return true
		}
	}

	return false
}

// GetExistingMarkers returns all codegen markers for a type.
func GetExistingMarkers(genDecl *ast.GenDecl) []ExistingMarker {
	if genDecl.Doc == nil {
		return nil
	}

	var markers []ExistingMarker

	for _, comment := range genDecl.Doc.List {
		text := strings.TrimSpace(comment.Text)
		// Remove comment markers
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		// Check if this is a codegen marker
		if strings.HasPrefix(text, TypeLevelPrefix) {
			content := strings.TrimPrefix(text, TypeLevelPrefix)
			content = strings.TrimSpace(content)

			// Parse generator type and config
			parts := strings.Fields(content)
			if len(parts) == 0 {
				continue
			}

			genType := parts[0]
			config := make(map[string]string)

			// Parse config key:value pairs
			for _, part := range parts[1:] {
				if idx := strings.Index(part, ":"); idx > 0 {
					key := part[:idx]
					value := part[idx+1:]
					config[key] = value
				}
			}

			markers = append(markers, ExistingMarker{
				GeneratorType: genType,
				Config:        config,
				OriginalText:  text,
			})
		}
	}

	return markers
}

// ExistingMarker represents a marker that's already present in the source.
type ExistingMarker struct {
	GeneratorType string
	Config        map[string]string
	OriginalText  string
}

// ShouldSkipMarking determines if we should skip marking based on existing markers.
// Returns true if the type already has the desired marker and force is false.
func ShouldSkipMarking(genDecl *ast.GenDecl, generatorType string, force bool) bool {
	if force {
		return false // Force mode means we always re-mark
	}

	return HasExistingMarker(genDecl, generatorType)
}

// FindMarkerComment finds the specific comment that contains a marker for the generator.
// Returns the index of the comment in the Doc.List, or -1 if not found.
func FindMarkerComment(genDecl *ast.GenDecl, generatorType string) int {
	if genDecl.Doc == nil {
		return -1
	}

	markerPrefix := TypeLevelPrefix + generatorType

	for i, comment := range genDecl.Doc.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		if strings.HasPrefix(text, markerPrefix) {
			return i
		}
	}

	return -1
}
