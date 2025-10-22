package tests

import (
	"testing"

	"github.com/ngicks/go-codegen/codegen/pkg/automark"
)

func TestIdempotentMarking(t *testing.T) {
	src := `package test

//codegen:cloner
type AlreadyMarked struct {
	Name string
}

type NotMarked struct {
	Value int
}
`

	// TODO: Test that:
	// 1. AlreadyMarked is detected as already having a marker
	// 2. Running automark again skips AlreadyMarked (unless --force)
	// 3. NotMarked gets a marker added
	// 4. Second run produces same result (idempotent)

	_ = src
}

func TestForceReMarking(t *testing.T) {
	src := `package test

//codegen:cloner
type Existing struct {
	Name string
}
`

	// TODO: Test that with --force flag:
	// 1. Existing marker is detected
	// 2. Old marker is replaced/updated with new config
	// 3. New marker reflects current command flags

	_ = src
}

func TestMultipleRunsStable(t *testing.T) {
	// Test that running automark multiple times produces stable output

	config := &automark.MarkerConfig{
		Generators: []automark.GeneratorSpec{
			{Type: automark.GeneratorCloner, Config: map[string]string{"no-copy": "copy"}},
		},
		DryRun: true,
	}

	// TODO: Test that:
	// 1. First run adds markers to eligible types
	// 2. Second run with same config makes no changes
	// 3. Output is byte-for-byte identical

	_ = config
}

func TestExistingMarkerDetection(t *testing.T) {
	tests := []struct {
		name       string
		comment    string
		genType    string
		shouldFind bool
	}{
		{
			name:       "exact match",
			comment:    "//codegen:cloner",
			genType:    automark.GeneratorCloner,
			shouldFind: true,
		},
		{
			name:       "with config",
			comment:    "//codegen:cloner no-copy:copy",
			genType:    automark.GeneratorCloner,
			shouldFind: true,
		},
		{
			name:       "different generator",
			comment:    "//codegen:und:patch",
			genType:    automark.GeneratorCloner,
			shouldFind: false,
		},
		{
			name:       "not a marker",
			comment:    "// This is just a comment",
			genType:    automark.GeneratorCloner,
			shouldFind: false,
		},
		{
			name:       "field directive not type marker",
			comment:    "//cloner:copyptr",
			genType:    automark.GeneratorCloner,
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Test HasExistingMarker function
			// This will check if a type already has a specific generator marker
			_ = tt
		})
	}
}

func TestConfigChangeDetection(t *testing.T) {
	// Test that changing config is detected and can trigger re-marking with --force

	oldMarker := "//codegen:cloner no-copy:ignore"
	newConfig := map[string]string{"no-copy": "copy", "chan": "make"}

	// TODO: Test that:
	// 1. Without --force, existing marker is preserved even if config differs
	// 2. With --force, marker is updated to reflect new config
	// 3. Config changes are correctly formatted in output

	_ = oldMarker
	_ = newConfig
}
