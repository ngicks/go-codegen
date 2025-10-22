package tests

import (
	"testing"

	"github.com/ngicks/go-codegen/codegen/pkg/automark"
	"gotest.tools/v3/assert"
)

func TestTypeNameFiltering(t *testing.T) {
	src := `package test

type UserRequest struct {
	Name string
}

type UserResponse struct {
	ID int
}

type InternalHelper struct {
	data string
}

type requestProcessor struct {
	count int
}
`

	tests := []struct {
		name         string
		include      []string
		exclude      []string
		exported     bool
		expectMarked []string
	}{
		{
			name:         "include pattern *Request",
			include:      []string{"*Request"},
			exclude:      nil,
			exported:     false,
			expectMarked: []string{"UserRequest"},
		},
		{
			name:         "include pattern *Response",
			include:      []string{"*Response"},
			exclude:      nil,
			exported:     false,
			expectMarked: []string{"UserResponse"},
		},
		{
			name:         "include both patterns",
			include:      []string{"*Request", "*Response"},
			exclude:      nil,
			exported:     false,
			expectMarked: []string{"UserRequest", "UserResponse"},
		},
		{
			name:         "exclude Internal*",
			include:      nil,
			exclude:      []string{"Internal*"},
			exported:     false,
			expectMarked: []string{"UserRequest", "UserResponse", "requestProcessor"},
		},
		{
			name:         "exported only",
			include:      nil,
			exclude:      nil,
			exported:     true,
			expectMarked: []string{"UserRequest", "UserResponse", "InternalHelper"},
		},
		{
			name:         "exclude takes precedence",
			include:      []string{"*Request", "*Response"},
			exclude:      []string{"*Response"},
			exported:     false,
			expectMarked: []string{"UserRequest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := automark.TypeFilterConfig{
				IncludeTypes:    tt.include,
				ExcludeTypes:    tt.exclude,
				RequireExported: tt.exported,
			}

			// TODO: Test that filter correctly selects types
			// Implementation will use ShouldMarkType(typeName, isExported, filter)
			_ = filter
			_ = src
		})
	}
}

func TestGlobPatternMatching(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		pattern     string
		shouldMatch bool
	}{
		{"exact match", "User", "User", true},
		{"prefix wildcard", "UserRequest", "*Request", true},
		{"suffix wildcard", "UserRequest", "User*", true},
		{"no match", "UserRequest", "*Response", false},
		{"middle wildcard", "UserServiceHandler", "User*Handler", true},
		{"multiple wildcards", "UserServiceHandler", "*Service*", true},
		{"question mark", "User", "Use?", true},
		{"question mark no match", "Users", "Use?", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Test glob matching function
			// Implementation will use matchGlob(typeName, pattern)
			_ = tt
		})
	}
}

func TestExportedTypeDetection(t *testing.T) {
	tests := []struct {
		name       string
		typeName   string
		isExported bool
	}{
		{"exported type", "User", true},
		{"unexported type", "user", false},
		{"exported with underscore", "User_Type", true},
		{"unexported with underscore", "user_type", false},
		{"single char exported", "U", true},
		{"single char unexported", "u", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Go's convention: exported = starts with uppercase
			exported := len(tt.typeName) > 0 && tt.typeName[0] >= 'A' && tt.typeName[0] <= 'Z'
			assert.Equal(t, tt.isExported, exported)
		})
	}
}

func TestFilterConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		filter      automark.TypeFilterConfig
		shouldError bool
	}{
		{
			name: "valid filter",
			filter: automark.TypeFilterConfig{
				IncludeTypes: []string{"*Request"},
				ExcludeTypes: []string{"Internal*"},
			},
			shouldError: false,
		},
		{
			name:        "empty filter is valid",
			filter:      automark.TypeFilterConfig{},
			shouldError: false,
		},
		{
			name: "include only",
			filter: automark.TypeFilterConfig{
				IncludeTypes: []string{"User*", "Admin*"},
			},
			shouldError: false,
		},
		{
			name: "exclude only",
			filter: automark.TypeFilterConfig{
				ExcludeTypes: []string{"test*"},
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Test ValidateFilterConfig function
			// Currently all filters are valid, but future may add validation
			_ = tt
		})
	}
}

func TestMultiplePatternMatching(t *testing.T) {
	typeName := "UserServiceRequest"

	patterns := []string{"*Request", "User*", "*Service*"}

	// Should match if ANY pattern matches
	for _, pattern := range patterns {
		// TODO: Test that matchGlob works correctly
		_ = pattern
	}

	_ = typeName
}
