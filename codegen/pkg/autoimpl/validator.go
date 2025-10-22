package autoimpl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateConfig validates the DispatchConfig before execution.
func ValidateConfig(cfg *DispatchConfig) error {
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

	// Validate generator registry
	if cfg.GeneratorRegistry == nil {
		return fmt.Errorf("generator registry is nil")
	}

	if len(cfg.GeneratorRegistry) == 0 {
		return fmt.Errorf("generator registry is empty")
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

	// Validate filter generators
	for _, gen := range cfg.OnlyGenerators {
		if !isKnownGenerator(gen) {
			return fmt.Errorf("unknown generator in --only: %s", gen)
		}
	}

	for _, gen := range cfg.SkipGenerators {
		if !isKnownGenerator(gen) {
			return fmt.Errorf("unknown generator in --skip: %s", gen)
		}
	}

	return nil
}

// isKnownGenerator checks if a generator type is known.
func isKnownGenerator(genType string) bool {
	known := []string{"cloner", "und:patch", "und:plain", "und:validator"}
	for _, g := range known {
		if g == genType {
			return true
		}
	}
	return false
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
