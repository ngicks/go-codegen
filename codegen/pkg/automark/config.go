package automark

// GeneratorSpec specifies which generator to run and with what configuration.
type GeneratorSpec struct {
	// Type is the generator type (e.g., "cloner", "und:patch", "und:plain")
	Type string

	// Config contains configuration options as key=value pairs
	Config map[string]string
}

// MarkerConfig is the configuration for the automark command execution.
type MarkerConfig struct {
	// PackagePatterns are the Go package patterns to process (e.g., "./...", "./pkg/models")
	PackagePatterns []string

	// Generators are the generators to mark types for
	Generators []GeneratorSpec

	// TypeFilters contains rules for which types to mark
	TypeFilters TypeFilterConfig

	// DryRun enables preview mode without writing files
	DryRun bool

	// Verbose enables verbose logging
	Verbose bool

	// WorkingDir is the base directory for operations
	WorkingDir string

	// Force overwrites existing markers
	Force bool
}

// TypeFilterConfig contains rules for determining which types are eligible for marking.
type TypeFilterConfig struct {
	// IncludeTypes are type name patterns to include (glob format)
	IncludeTypes []string

	// ExcludeTypes are type name patterns to exclude (glob format)
	ExcludeTypes []string

	// RequireExported only marks exported (capitalized) types
	RequireExported bool

	// MatcherRules reuses existing matcher logic for type eligibility
	// This will be populated with generator-specific matching rules
	MatcherRules interface{}
}
