package directive

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseDirectiveLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "simple flags",
			input: "ignore generated",
			expected: map[string]string{
				"ignore":    "",
				"generated": "",
			},
		},
		{
			name:  "key:value pairs",
			input: "name:foo count:42",
			expected: map[string]string{
				"name":  "foo",
				"count": "42",
			},
		},
		{
			name:  "quoted values with spaces",
			input: `message:"hello world" path:"/some/path with spaces"`,
			expected: map[string]string{
				"message": "hello world",
				"path":    "/some/path with spaces",
			},
		},
		{
			name:  "raw string literals",
			input: "regex:`\\d+` template:`{{.Name}}`",
			expected: map[string]string{
				"regex":    `\d+`,
				"template": `{{.Name}}`,
			},
		},
		{
			name:  "mixed formats",
			input: `flag name:simple quoted:"with spaces" raw:` + "`raw`",
			expected: map[string]string{
				"flag":   "",
				"name":   "simple",
				"quoted": "with spaces",
				"raw":    "raw",
			},
		},
		{
			name:  "escaped quotes",
			input: `text:"He said \"hello\""`,
			expected: map[string]string{
				"text": `He said "hello"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDirectiveLine(tt.input)
			assert.NilError(t, err)
			assert.DeepEqual(t, tt.expected, result)
		})
	}
}

// TestUnmarshalDirectiveInterface tests the UnmarshalDirective interface
type CustomDirective struct {
	Name    string
	Count   int
	Enabled bool
}

func (c *CustomDirective) UnmarshalDirective(m map[string]string) error {
	if name, ok := m["name"]; ok {
		c.Name = name
	}
	if _, ok := m["enabled"]; ok {
		c.Enabled = true
	}
	// Custom logic for count - defaults to 10 if not specified
	c.Count = 10
	if countStr, ok := m["count"]; ok && countStr != "" {
		// In real implementation, would parse the int
		c.Count = 42 // simplified for test
	}
	return nil
}

func TestUnmarshalWithInterface(t *testing.T) {
	parsed := ParsedDirective{
		"name":    "test",
		"enabled": "",
		"count":   "42",
	}

	var custom CustomDirective
	err := parsed.Unmarshal(&custom)
	assert.NilError(t, err)
	assert.Equal(t, "test", custom.Name)
	assert.Equal(t, true, custom.Enabled)
	assert.Equal(t, 42, custom.Count)
}