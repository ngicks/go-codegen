package autoimpl

import "golang.org/x/tools/go/packages"

// DispatchConfig is the configuration for the autoimpl command execution.
type DispatchConfig struct {
	// PackagePatterns are the Go package patterns to scan for marked types
	PackagePatterns []string

	// GeneratorRegistry maps generator type strings to generator functions
	GeneratorRegistry GeneratorRegistry

	// Verbose enables verbose logging
	Verbose bool

	// DryRun enables preview mode without writing files
	DryRun bool

	// IgnoreGenerated skips scanning generated files for markers
	IgnoreGenerated bool

	// WorkingDir is the base directory for operations
	WorkingDir string

	// OnlyGenerators filters to only run these generators
	OnlyGenerators []string

	// SkipGenerators skips these generators even if marked
	SkipGenerators []string
}

// GeneratorFunc is the signature for generator functions.
// It receives marked types, package information, and configuration.
type GeneratorFunc func(marked []MarkedType, pkgs []*packages.Package, config map[string]string) error

// GeneratorRegistry maps generator type strings to executable generator functions.
type GeneratorRegistry map[string]GeneratorFunc

// MarkedType represents a type that has been marked for code generation.
type MarkedType struct {
	// TypeName is the name of the type
	TypeName string

	// Package is the package containing the type
	Package *packages.Package

	// Generators are the list of generators to apply (parsed from markers)
	Generators []string

	// Config contains parsed configuration from marker comments
	Config map[string]string

	// FilePath is the source file path
	FilePath string

	// Line is the line number where the type is declared
	Line int
}
