package automark

// Directive comment format constants
const (
	// TypeLevelPrefix is the prefix for type-level markers
	// Format: //codegen:generatortype config
	TypeLevelPrefix = "codegen:"

	// GeneratorCloner is the cloner generator identifier
	GeneratorCloner = "cloner"

	// GeneratorUndPatch is the und:patch generator identifier
	GeneratorUndPatch = "und:patch"

	// GeneratorUndPlain is the und:plain generator identifier
	GeneratorUndPlain = "und:plain"

	// GeneratorUndValidator is the und:validator generator identifier
	GeneratorUndValidator = "und:validator"
)

// IsValidGenerator checks if the given generator type is recognized.
func IsValidGenerator(genType string) bool {
	switch genType {
	case GeneratorCloner, GeneratorUndPatch, GeneratorUndPlain, GeneratorUndValidator:
		return true
	default:
		return false
	}
}

// AllValidGenerators returns a slice of all valid generator types.
func AllValidGenerators() []string {
	return []string{
		GeneratorCloner,
		GeneratorUndPatch,
		GeneratorUndPlain,
		GeneratorUndValidator,
	}
}
