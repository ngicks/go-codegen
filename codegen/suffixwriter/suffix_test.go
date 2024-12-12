package suffixwriter

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_suffix(t *testing.T) {
	for _, set := range [][2]string{
		{"foo.go", "foo.suffix.go"},
		{"foo_test.go", "foo.suffix_test.go"},
		{"foo_linux.go", "foo.suffix_linux.go"},
		{"foo_amd64.go", "foo.suffix_amd64.go"},
		{"foo_linux_amd64.go", "foo.suffix_linux_amd64.go"},
		{"foo_linux_amd64_test.go", "foo.suffix_linux_amd64_test.go"},
		{"foo_amd64_linux.go", "foo_amd64.suffix_linux.go"}, // wrong
		{"foo_bar_linux.go", "foo_bar.suffix_linux.go"},
		{"foo_bar.go", "foo_bar.suffix.go"},
	} {
		assert.Equal(t, set[1], suffixFilename(set[0], ".suffix"))
	}
}
