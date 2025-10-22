package tests

import (
	"testing"

	"github.com/ngicks/go-codegen/codegen/pkg/autoimpl"
)

func TestDispatchGrouping(t *testing.T) {
	// Test that dispatch correctly groups types by generator

	marked := []autoimpl.MarkedType{
		{TypeName: "Type1", Generators: []string{"cloner"}},
		{TypeName: "Type2", Generators: []string{"cloner"}},
		{TypeName: "Type3", Generators: []string{"und:patch"}},
		{TypeName: "Type4", Generators: []string{"cloner", "und:patch"}},
	}

	// TODO: Test that Dispatch groups correctly:
	// - cloner: Type1, Type2, Type4
	// - und:patch: Type3, Type4

	_ = marked
}

func TestDispatchFiltering(t *testing.T) {
	marked := []autoimpl.MarkedType{
		{TypeName: "Type1", Generators: []string{"cloner", "und:patch", "und:plain"}},
	}

	cfg := &autoimpl.DispatchConfig{
		OnlyGenerators: []string{"cloner"},
	}

	// TODO: Test that --only flag filters correctly
	// Should only dispatch to cloner, skip und generators

	_ = marked
	_ = cfg
}

func TestDispatchSkipping(t *testing.T) {
	marked := []autoimpl.MarkedType{
		{TypeName: "Type1", Generators: []string{"cloner", "und:patch", "und:plain"}},
	}

	cfg := &autoimpl.DispatchConfig{
		SkipGenerators: []string{"und:patch"},
	}

	// TODO: Test that --skip flag works correctly
	// Should dispatch to cloner and und:plain, skip und:patch

	_ = marked
	_ = cfg
}

func TestDispatchPrecedence(t *testing.T) {
	// Test that --skip takes precedence over --only

	marked := []autoimpl.MarkedType{
		{TypeName: "Type1", Generators: []string{"cloner", "und:patch"}},
	}

	cfg := &autoimpl.DispatchConfig{
		OnlyGenerators: []string{"cloner", "und:patch"},
		SkipGenerators: []string{"cloner"},
	}

	// TODO: Test precedence
	// Should only dispatch to und:patch (skip overrides only)

	_ = marked
	_ = cfg
}

func TestUnknownGeneratorError(t *testing.T) {
	marked := []autoimpl.MarkedType{
		{TypeName: "Type1", Generators: []string{"unknown"}},
	}

	registry := autoimpl.DefaultRegistry()
	cfg := &autoimpl.DispatchConfig{
		GeneratorRegistry: registry,
	}

	// TODO: Test that unknown generator type produces error
	// Should return error: "unknown generator type: unknown"

	_ = marked
	_ = cfg
}

func TestDispatchWithEmptyMarked(t *testing.T) {
	marked := []autoimpl.MarkedType{}

	cfg := &autoimpl.DispatchConfig{
		GeneratorRegistry: autoimpl.DefaultRegistry(),
	}

	// TODO: Test that empty marked list completes successfully
	// Should not error, should not invoke any generators

	_ = marked
	_ = cfg
}

func TestDispatchErrorHandling(t *testing.T) {
	// Test that generator errors are properly propagated

	marked := []autoimpl.MarkedType{
		{TypeName: "Type1", Generators: []string{"cloner"}},
	}

	// TODO: Test error handling
	// If a generator fails, Dispatch should return that error
	// Error message should include generator name and type name

	_ = marked
}

func TestConfigPassthrough(t *testing.T) {
	// Test that marker config is passed to generator functions

	config := map[string]string{
		"no-copy": "copy",
		"chan":    "make",
	}

	marked := []autoimpl.MarkedType{
		{TypeName: "Type1", Generators: []string{"cloner"}, Config: config},
	}

	// TODO: Test that config is correctly passed to generator function
	// Generator should receive the config map

	_ = marked
}
