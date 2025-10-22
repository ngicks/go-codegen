package tests

import (
	"testing"

	"github.com/ngicks/go-codegen/codegen/pkg/autoimpl"
	"gotest.tools/v3/assert"
)

func TestMarkerParsing(t *testing.T) {
	tests := []struct {
		name           string
		markerText     string
		expectedGen    string
		expectedConfig map[string]string
		shouldError    bool
	}{
		{
			name:           "simple cloner marker",
			markerText:     "codegen:cloner",
			expectedGen:    "cloner",
			expectedConfig: map[string]string{},
			shouldError:    false,
		},
		{
			name:        "cloner with config",
			markerText:  "codegen:cloner no-copy:copy chan:make",
			expectedGen: "cloner",
			expectedConfig: map[string]string{
				"no-copy": "copy",
				"chan":    "make",
			},
			shouldError: false,
		},
		{
			name:           "und:patch marker",
			markerText:     "codegen:und:patch",
			expectedGen:    "und:patch",
			expectedConfig: map[string]string{},
			shouldError:    false,
		},
		{
			name:           "und:plain marker",
			markerText:     "codegen:und:plain",
			expectedGen:    "und:plain",
			expectedConfig: map[string]string{},
			shouldError:    false,
		},
		{
			name:           "und:validator marker",
			markerText:     "codegen:und:validator",
			expectedGen:    "und:validator",
			expectedConfig: map[string]string{},
			shouldError:    false,
		},
		{
			name:           "invalid marker (no codegen prefix)",
			markerText:     "cloner",
			expectedGen:    "",
			expectedConfig: nil,
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Test ParseMarker function
			// This will parse a marker comment and extract generator type and config
			_ = tt
		})
	}
}

func TestMultipleMarkerParsing(t *testing.T) {
	src := `package test

//codegen:cloner no-copy:copy
//codegen:und:patch
//codegen:und:plain
type MultiMarked struct {
	Name string
}
`

	// TODO: Test that ParseTypeMarkers extracts all three markers
	// Expected result:
	// - cloner with config {no-copy: copy}
	// - und:patch with no config
	// - und:plain with no config

	_ = src
}

func TestMarkerCommentFormat(t *testing.T) {
	// Test various comment formats
	formats := []string{
		"// codegen:cloner",
		"//codegen:cloner",
		"//  codegen:cloner  ",
		"/* codegen:cloner */",
	}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			// TODO: Test that all formats are correctly parsed
			// Should normalize whitespace and handle both // and /* */ comments
			_ = format
		})
	}
}

func TestConfigKeyValueParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "single key:value",
			input: "no-copy:copy",
			expected: map[string]string{
				"no-copy": "copy",
			},
		},
		{
			name:  "multiple key:value pairs",
			input: "no-copy:copy chan:make func:ignore",
			expected: map[string]string{
				"no-copy": "copy",
				"chan":    "make",
				"func":    "ignore",
			},
		},
		{
			name:  "with extra whitespace",
			input: "no-copy:copy  chan:make   func:ignore",
			expected: map[string]string{
				"no-copy": "copy",
				"chan":    "make",
				"func":    "ignore",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Test parseConfig function
			_ = tt
		})
	}
}

func TestInvalidMarkerDetection(t *testing.T) {
	invalid := []string{
		"",                // Empty
		"codegen:",        // No generator type
		"invalid:cloner",  // Wrong prefix
		"codegen:unknown", // Unknown generator
		"codegen cloner",  // Missing colon
	}

	for _, marker := range invalid {
		t.Run(marker, func(t *testing.T) {
			// TODO: Test that invalid markers are rejected
			// Should return error or nil result
			_ = marker
		})
	}
}

func TestMarkerScanningInPackage(t *testing.T) {
	// Test scanning a package for all marked types

	src := `package test

//codegen:cloner
type Marked1 struct {
	Name string
}

type Unmarked struct {
	Value int
}

//codegen:und:patch
type Marked2 struct {
	Data []byte
}
`

	// TODO: Test ScanForMarkedTypes function
	// Expected: finds Marked1 and Marked2, skips Unmarked
	// Returns: []MarkedType with correct type names and generators

	_ = src
}

func TestRegistryDefaultValues(t *testing.T) {
	registry := autoimpl.DefaultRegistry()

	// Should have all four generators
	assert.Equal(t, 4, len(registry))

	// Check each generator exists
	_, ok := registry["cloner"]
	assert.Assert(t, ok, "cloner generator missing")

	_, ok = registry["und:patch"]
	assert.Assert(t, ok, "und:patch generator missing")

	_, ok = registry["und:plain"]
	assert.Assert(t, ok, "und:plain generator missing")

	_, ok = registry["und:validator"]
	assert.Assert(t, ok, "und:validator generator missing")
}
