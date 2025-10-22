package autoimpl

import (
	"go/ast"
	"strings"
)

const (
	// MarkerPrefix is the prefix for type-level codegen markers
	MarkerPrefix = "codegen:"
)

// ParsedMarker represents a parsed codegen marker.
type ParsedMarker struct {
	GeneratorType string
	Config        map[string]string
	OriginalText  string
}

// parseTypeMarkers parses all codegen markers from a comment group.
func parseTypeMarkers(comments *ast.CommentGroup) []ParsedMarker {
	if comments == nil || len(comments.List) == 0 {
		return nil
	}

	var markers []ParsedMarker

	for _, comment := range comments.List {
		text := strings.TrimSpace(comment.Text)

		// Remove comment markers
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)

		// Check if this is a codegen marker
		if !strings.HasPrefix(text, MarkerPrefix) {
			continue
		}

		// Parse the marker
		marker := parseMarker(text)
		if marker != nil {
			markers = append(markers, *marker)
		}
	}

	return markers
}

// parseMarker parses a single codegen marker comment.
// Input format: "codegen:generatortype key1:value1 key2:value2"
func parseMarker(text string) *ParsedMarker {
	// Remove prefix
	if !strings.HasPrefix(text, MarkerPrefix) {
		return nil
	}

	content := strings.TrimPrefix(text, MarkerPrefix)
	content = strings.TrimSpace(content)

	if content == "" {
		return nil
	}

	// Split into parts
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return nil
	}

	// First part is generator type
	generatorType := parts[0]

	// Parse config from remaining parts
	config := make(map[string]string)
	for _, part := range parts[1:] {
		idx := strings.Index(part, ":")
		if idx > 0 {
			key := part[:idx]
			value := part[idx+1:]
			config[key] = value
		} else if part != "" {
			// Boolean flag (no value)
			config[part] = ""
		}
	}

	return &ParsedMarker{
		GeneratorType: generatorType,
		Config:        config,
		OriginalText:  text,
	}
}

// ValidateMarker validates a parsed marker.
func ValidateMarker(marker ParsedMarker) error {
	// Check if generator type is recognized
	validGenerators := map[string]bool{
		"cloner":        true,
		"und:patch":     true,
		"und:plain":     true,
		"und:validator": true,
	}

	if !validGenerators[marker.GeneratorType] {
		return &InvalidMarkerError{
			GeneratorType: marker.GeneratorType,
			Message:       "unknown generator type",
		}
	}

	// Validate config based on generator type
	switch marker.GeneratorType {
	case "cloner":
		return validateClonerConfig(marker.Config)
	case "und:patch", "und:plain", "und:validator":
		// Und generators don't have config options currently
		if len(marker.Config) > 0 {
			return &InvalidMarkerError{
				GeneratorType: marker.GeneratorType,
				Message:       "generator does not support configuration options",
			}
		}
	}

	return nil
}

// validateClonerConfig validates cloner generator configuration.
func validateClonerConfig(config map[string]string) error {
	validKeys := map[string][]string{
		"no-copy":   {"ignore", "disallow", "copy"},
		"chan":      {"ignore", "disallow", "copy", "make"},
		"func":      {"ignore", "disallow", "copy"},
		"interface": {"ignore", "copy"},
	}

	for key, value := range config {
		validValues, ok := validKeys[key]
		if !ok {
			return &InvalidMarkerError{
				GeneratorType: "cloner",
				Message:       "unknown config key: " + key,
			}
		}

		if value != "" {
			found := false
			for _, v := range validValues {
				if value == v {
					found = true
					break
				}
			}
			if !found {
				return &InvalidMarkerError{
					GeneratorType: "cloner",
					Message:       "invalid value for " + key + ": " + value,
				}
			}
		}
	}

	return nil
}

// InvalidMarkerError represents an error parsing or validating a marker.
type InvalidMarkerError struct {
	GeneratorType string
	Message       string
	FilePath      string
	Line          int
}

func (e *InvalidMarkerError) Error() string {
	if e.FilePath != "" {
		return "invalid marker at " + e.FilePath + ":" + string(rune(e.Line)) + ": " + e.Message
	}
	return "invalid marker for " + e.GeneratorType + ": " + e.Message
}
