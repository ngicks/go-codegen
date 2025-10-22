package automark_autoimpl_test

import (
	"context"
	"os"
	"testing"

	"github.com/ngicks/go-codegen/codegen/pkg/automark"
	"github.com/ngicks/go-codegen/codegen/pkg/autoimpl"
	"gotest.tools/v3/assert"
)

func TestEndToEndWorkflow(t *testing.T) {
	// This test validates the complete automark → autoimpl workflow

	// Setup: Use the test input directory
	testDir := "./input"
	ctx := context.Background()

	// Get current working directory for test
	wd, err := os.Getwd()
	assert.NilError(t, err, "should get working directory")

	// Step 1: Run automark to mark types
	markCfg := &automark.MarkerConfig{
		PackagePatterns: []string{testDir},
		Generators: []automark.GeneratorSpec{
			{Type: automark.GeneratorCloner, Config: map[string]string{}},
		},
		TypeFilters: automark.TypeFilterConfig{},
		DryRun:      true, // Dry run to not modify test files
		Verbose:     testing.Verbose(),
		WorkingDir:  wd,
		Force:       false,
	}

	result, err := automark.Mark(ctx, markCfg)
	assert.NilError(t, err, "automark should complete without error")
	assert.Assert(t, result != nil, "automark should return result")

	if testing.Verbose() {
		t.Logf("Automark discovered %d types", result.TotalDiscovered)
		t.Logf("Automark filtered to %d types", result.TotalFiltered)
	}

	// Step 2: Run autoimpl to scan for markers
	// Note: Since we're in dry-run mode for automark, we won't have actual markers
	// This test validates the structure works, not end-to-end file modification

	implCfg := &autoimpl.DispatchConfig{
		PackagePatterns:   []string{testDir},
		GeneratorRegistry: autoimpl.DefaultRegistry(),
		Verbose:           testing.Verbose(),
		DryRun:            true,
		IgnoreGenerated:   false,
		WorkingDir:        wd,
	}

	err = autoimpl.ValidateConfig(implCfg)
	assert.NilError(t, err, "autoimpl config should be valid")

	// TODO: Once full implementation is complete, scan for marked types
	// marked, err := autoimpl.ScanForMarkedTypes(ctx, implCfg)
	// assert.NilError(t, err, "autoimpl should scan without error")
}

func TestAutomarkDryRun(t *testing.T) {
	// Test that dry-run mode doesn't modify files

	testDir := "./input"
	ctx := context.Background()

	// Get current working directory for test
	wd, err := os.Getwd()
	assert.NilError(t, err, "should get working directory")

	// Read directory before
	beforeFiles, err := os.ReadDir(testDir)
	assert.NilError(t, err)

	cfg := &automark.MarkerConfig{
		PackagePatterns: []string{testDir},
		Generators: []automark.GeneratorSpec{
			{Type: automark.GeneratorCloner, Config: map[string]string{}},
		},
		DryRun:     true,
		Verbose:    testing.Verbose(),
		WorkingDir: wd,
	}

	_, err = automark.Mark(ctx, cfg)
	assert.NilError(t, err)

	// Read directory after
	afterFiles, err := os.ReadDir(testDir)
	assert.NilError(t, err)

	// Should have same number of files
	assert.Equal(t, len(beforeFiles), len(afterFiles), "dry-run should not create/delete files")
}

func TestValidation(t *testing.T) {
	// Test configuration validation

	t.Run("automark validation", func(t *testing.T) {
		// Valid config
		valid := &automark.MarkerConfig{
			PackagePatterns: []string{"./..."},
			Generators: []automark.GeneratorSpec{
				{Type: "cloner", Config: map[string]string{}},
			},
		}
		err := automark.ValidateConfig(valid)
		assert.NilError(t, err, "valid config should pass")

		// Invalid: no packages
		invalid := &automark.MarkerConfig{
			PackagePatterns: []string{},
			Generators: []automark.GeneratorSpec{
				{Type: "cloner", Config: map[string]string{}},
			},
		}
		err = automark.ValidateConfig(invalid)
		assert.Assert(t, err != nil, "config with no packages should fail")

		// Invalid: no generators
		invalid2 := &automark.MarkerConfig{
			PackagePatterns: []string{"./..."},
			Generators:      []automark.GeneratorSpec{},
		}
		err = automark.ValidateConfig(invalid2)
		assert.Assert(t, err != nil, "config with no generators should fail")
	})

	t.Run("autoimpl validation", func(t *testing.T) {
		// Valid config
		valid := &autoimpl.DispatchConfig{
			PackagePatterns:   []string{"./..."},
			GeneratorRegistry: autoimpl.DefaultRegistry(),
		}
		err := autoimpl.ValidateConfig(valid)
		assert.NilError(t, err, "valid config should pass")

		// Invalid: no packages
		invalid := &autoimpl.DispatchConfig{
			PackagePatterns:   []string{},
			GeneratorRegistry: autoimpl.DefaultRegistry(),
		}
		err = autoimpl.ValidateConfig(invalid)
		assert.Assert(t, err != nil, "config with no packages should fail")

		// Invalid: no registry
		invalid2 := &autoimpl.DispatchConfig{
			PackagePatterns:   []string{"./..."},
			GeneratorRegistry: nil,
		}
		err = autoimpl.ValidateConfig(invalid2)
		assert.Assert(t, err != nil, "config with no registry should fail")
	})
}
