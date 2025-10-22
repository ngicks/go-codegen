package automark

import (
	"path/filepath"
	"strings"
)

// ApplyFilters filters type candidates based on the configured filter rules.
func ApplyFilters(candidates []TypeCandidate, filter TypeFilterConfig) []TypeCandidate {
	var filtered []TypeCandidate

	for _, candidate := range candidates {
		if ShouldMarkType(candidate, filter) {
			filtered = append(filtered, candidate)
		}
	}

	return filtered
}

// ShouldMarkType determines if a type candidate should be marked based on filter rules.
func ShouldMarkType(candidate TypeCandidate, filter TypeFilterConfig) bool {
	typeName := candidate.TypeName

	// Check exported-only filter
	if filter.RequireExported && !candidate.IsExported {
		return false
	}

	// Check exclude patterns (takes precedence)
	for _, pattern := range filter.ExcludeTypes {
		if matchGlob(typeName, pattern) {
			return false
		}
	}

	// Check include patterns
	if len(filter.IncludeTypes) > 0 {
		matched := false
		for _, pattern := range filter.IncludeTypes {
			if matchGlob(typeName, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// All filters passed
	return true
}

// matchGlob performs glob pattern matching for type names.
// Supports * (matches any sequence) and ? (matches single character).
func matchGlob(name, pattern string) bool {
	// Use filepath.Match which supports glob patterns
	// This handles *, ?, and [...] patterns
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		// Invalid pattern - treat as no match
		return false
	}
	return matched
}

// FormatFilterSummary creates a human-readable summary of filter rules.
func FormatFilterSummary(filter TypeFilterConfig) string {
	var parts []string

	if filter.RequireExported {
		parts = append(parts, "exported types only")
	}

	if len(filter.IncludeTypes) > 0 {
		parts = append(parts, "include: "+strings.Join(filter.IncludeTypes, ", "))
	}

	if len(filter.ExcludeTypes) > 0 {
		parts = append(parts, "exclude: "+strings.Join(filter.ExcludeTypes, ", "))
	}

	if len(parts) == 0 {
		return "no filters (all types eligible)"
	}

	return strings.Join(parts, "; ")
}
