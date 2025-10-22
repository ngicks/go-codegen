package automark

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateConfig validates the MarkerConfig before execution.
func ValidateConfig(cfg *MarkerConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Validate package patterns
	if len(cfg.PackagePatterns) == 0 {
		return fmt.Errorf("no package patterns specified")
	}

	for _, pattern := range cfg.PackagePatterns {
		if pattern == "" {
			return fmt.Errorf("empty package pattern")
		}
		// Package patterns should be relative (security constraint)
		if filepath.IsAbs(pattern) {
			return fmt.Errorf("package pattern must be relative, got: %s", pattern)
		}
	}

	// Validate generators
	if len(cfg.Generators) == 0 {
		return fmt.Errorf("no generators specified")
	}

	for i, gen := range cfg.Generators {
		if err := ValidateGeneratorSpec(gen); err != nil {
			return fmt.Errorf("generator %d: %w", i, err)
		}
	}

	// Validate working directory
	if cfg.WorkingDir != "" {
		if !filepath.IsAbs(cfg.WorkingDir) {
			// Convert to absolute
			abs, err := filepath.Abs(cfg.WorkingDir)
			if err != nil {
				return fmt.Errorf("invalid working directory %q: %w", cfg.WorkingDir, err)
			}
			cfg.WorkingDir = abs
		}

		// Check it exists
		stat, err := os.Stat(cfg.WorkingDir)
		if err != nil {
			return fmt.Errorf("working directory %q: %w", cfg.WorkingDir, err)
		}
		if !stat.IsDir() {
			return fmt.Errorf("working directory %q is not a directory", cfg.WorkingDir)
		}
	} else {
		// Use current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		cfg.WorkingDir = cwd
	}

	// Validate filter config
	if err := ValidateTypeFilterConfig(cfg.TypeFilters); err != nil {
		return fmt.Errorf("invalid type filters: %w", err)
	}

	return nil
}

// ValidateGeneratorSpec validates a single GeneratorSpec.
func ValidateGeneratorSpec(spec GeneratorSpec) error {
	if spec.Type == "" {
		return fmt.Errorf("generator type is empty")
	}

	if !IsValidGenerator(spec.Type) {
		return fmt.Errorf("unknown generator type %q, valid types: %s",
			spec.Type, strings.Join(AllValidGenerators(), ", "))
	}

	// Validate config keys based on generator type
	switch spec.Type {
	case GeneratorCloner:
		validKeys := map[string][]string{
			"no-copy":   {"ignore", "disallow", "copy"},
			"chan":      {"ignore", "disallow", "copy", "make"},
			"func":      {"ignore", "disallow", "copy"},
			"interface": {"ignore", "copy"},
		}

		for key, value := range spec.Config {
			validValues, ok := validKeys[key]
			if !ok {
				return fmt.Errorf("unknown config key %q for cloner generator", key)
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
					return fmt.Errorf("invalid value %q for key %q, valid values: %s",
						value, key, strings.Join(validValues, ", "))
				}
			}
		}

	case GeneratorUndPatch, GeneratorUndPlain, GeneratorUndValidator:
		// Und generators currently don't have configuration options
		if len(spec.Config) > 0 {
			return fmt.Errorf("%s generator does not support configuration options", spec.Type)
		}
	}

	return nil
}

// ValidateTypeFilterConfig validates a TypeFilterConfig.
func ValidateTypeFilterConfig(filter TypeFilterConfig) error {
	// Validate glob patterns
	for _, pattern := range filter.IncludeTypes {
		if !isValidGlobPattern(pattern) {
			return fmt.Errorf("invalid include pattern: %q", pattern)
		}
	}

	for _, pattern := range filter.ExcludeTypes {
		if !isValidGlobPattern(pattern) {
			return fmt.Errorf("invalid exclude pattern: %q", pattern)
		}
	}

	return nil
}

// isValidGlobPattern checks if a pattern is a valid glob pattern.
func isValidGlobPattern(pattern string) bool {
	if pattern == "" {
		return false
	}

	// Try to match against a dummy string to validate the pattern
	_, err := filepath.Match(pattern, "test")
	return err == nil
}

// ValidatePaths ensures all paths are within the working directory (security check).
func ValidatePaths(workingDir string, paths []string) error {
	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("getting absolute working directory: %w", err)
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("getting absolute path for %q: %w", path, err)
		}

		// Check if path is under working directory
		relPath, err := filepath.Rel(absWorkingDir, absPath)
		if err != nil {
			return fmt.Errorf("getting relative path: %w", err)
		}

		// Path traversal check
		if strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("path %q is outside working directory %q", path, workingDir)
		}
	}

	return nil
}
