package tests

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/ngicks/go-codegen/codegen/pkg/automark"
	"gotest.tools/v3/assert"
)

func TestBasicTypeMarking(t *testing.T) {
	src := `package test

type SimpleStruct struct {
	Name string
	Age  int
}

type GenericStruct[T any] struct {
	Value T
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	assert.NilError(t, err)

	// Test that we can add markers to types
	spec := automark.GeneratorSpec{
		Type:   automark.GeneratorCloner,
		Config: map[string]string{},
	}

	// Find SimpleStruct type
	var simpleType *ast.TypeSpec
	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == "SimpleStruct" {
			simpleType = ts
			return false
		}
		return true
	})

	assert.Assert(t, simpleType != nil, "SimpleStruct not found")

	// TODO: Test marker addition logic
	// This is a placeholder - actual implementation will test:
	// 1. Marker is added above type declaration
	// 2. Marker format is correct (//codegen:cloner)
	// 3. Existing comments are preserved
	_ = spec // placeholder for future test implementation
}

func TestMarkerFormat(t *testing.T) {
	tests := []struct {
		name     string
		genType  string
		config   map[string]string
		expected string
	}{
		{
			name:     "simple cloner",
			genType:  automark.GeneratorCloner,
			config:   map[string]string{},
			expected: "//codegen:cloner",
		},
		{
			name:    "cloner with config",
			genType: automark.GeneratorCloner,
			config: map[string]string{
				"no-copy": "copy",
				"chan":    "make",
			},
			expected: "//codegen:cloner no-copy:copy chan:make",
		},
		{
			name:     "und:patch",
			genType:  automark.GeneratorUndPatch,
			config:   map[string]string{},
			expected: "//codegen:und:patch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Test that FormatMarker produces correct output
			// This will be implemented when we add the FormatMarker function
			_ = tt.expected
		})
	}
}

func TestExistingCommentPreservation(t *testing.T) {
	src := `package test

// SimpleStruct is a test struct
// with multi-line documentation
type SimpleStruct struct {
	Name string
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	assert.NilError(t, err)

	// Find SimpleStruct
	var simpleType *ast.TypeSpec
	var doc *ast.CommentGroup
	ast.Inspect(file, func(n ast.Node) bool {
		if gd, ok := n.(*ast.GenDecl); ok {
			if len(gd.Specs) > 0 {
				if ts, ok := gd.Specs[0].(*ast.TypeSpec); ok && ts.Name.Name == "SimpleStruct" {
					simpleType = ts
					doc = gd.Doc
					return false
				}
			}
		}
		return true
	})

	assert.Assert(t, simpleType != nil, "SimpleStruct not found")
	assert.Assert(t, doc != nil, "Documentation not found")
	assert.Assert(t, len(doc.List) > 0, "Expected documentation comments")

	// TODO: Test that adding marker preserves existing documentation
	// Marker should be added before the doc comments
}

func TestMultipleGenerators(t *testing.T) {
	// Test that a type can be marked for multiple generators
	specs := []automark.GeneratorSpec{
		{Type: automark.GeneratorCloner, Config: map[string]string{}},
		{Type: automark.GeneratorUndPatch, Config: map[string]string{}},
		{Type: automark.GeneratorUndPlain, Config: map[string]string{}},
	}

	// TODO: Test that all three markers are added
	// Expected output:
	// //codegen:cloner
	// //codegen:und:patch
	// //codegen:und:plain
	// type Foo struct { ... }

	_ = specs
}

func TestValidatorFunctions(t *testing.T) {
	tests := []struct {
		name      string
		genType   string
		wantValid bool
	}{
		{"cloner is valid", automark.GeneratorCloner, true},
		{"und:patch is valid", automark.GeneratorUndPatch, true},
		{"und:plain is valid", automark.GeneratorUndPlain, true},
		{"und:validator is valid", automark.GeneratorUndValidator, true},
		{"invalid type", "invalid", false},
		{"empty type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := automark.IsValidGenerator(tt.genType)
			assert.Equal(t, tt.wantValid, valid)
		})
	}
}

func TestAllValidGenerators(t *testing.T) {
	gens := automark.AllValidGenerators()
	assert.Equal(t, 4, len(gens))
	assert.Assert(t, contains(gens, automark.GeneratorCloner))
	assert.Assert(t, contains(gens, automark.GeneratorUndPatch))
	assert.Assert(t, contains(gens, automark.GeneratorUndPlain))
	assert.Assert(t, contains(gens, automark.GeneratorUndValidator))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hasPrefix(comments *ast.CommentGroup, prefix string) bool {
	if comments == nil {
		return false
	}
	for _, c := range comments.List {
		if strings.HasPrefix(strings.TrimSpace(c.Text[2:]), prefix) {
			return true
		}
	}
	return false
}
