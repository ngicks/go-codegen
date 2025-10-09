package directive

import (
	"fmt"
	"go/ast"
	"iter"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/ngicks/go-iterator-helper/hiter"
)

// UnmarshalDirective is an interface for custom directive unmarshaling
type UnmarshalDirective interface {
	UnmarshalDirective(m map[string]string) error
}

// ParsedDirective represents parsed directive data as key-value pairs
type ParsedDirective map[string]string

// ParseDirectiveGeneric parses directive comments into key=value pairs
// Format: //prefix:key1=value1 key2="value with spaces" key3 ...
// For simple directives without '=', the value is empty string
// Values can be quoted using strconv.Quote rules to include whitespace
func Parse(seq iter.Seq[string], prefix string) (ParsedDirective, bool) {
	lines := extractDirectiveLines(seq, prefix, true)
	if len(lines) == 0 {
		return nil, false
	}

	result := make(ParsedDirective)
	for _, line := range lines {
		// Parse the line for key:value pairs
		parts, err := parseDirectiveLine(line)
		if err != nil {
			// If parsing fails, skip this line
			continue
		}
		for k, v := range parts {
			result[k] = v
		}
	}

	return result, true
}

// ParseAst parses directive comments from an AST CommentGroup
func ParseAst(comments *ast.CommentGroup, prefix string) (ParsedDirective, bool) {
	return Parse(EnumerateCommentGroup(comments), prefix)
}

// parseDirectiveLine parses a single directive line into key=value pairs
func parseDirectiveLine(line string) (map[string]string, error) {
	result := make(map[string]string)
	remaining := line

	for remaining != "" {
		remaining = strings.TrimSpace(remaining)
		if remaining == "" {
			break
		}

		// Find the next key
		var key, value string

		// Check if there's a ':' sign
		colonIdx := strings.IndexByte(remaining, ':')

		// Find the end of the key (either ':' or whitespace)
		keyEnd := len(remaining)
		for i, r := range remaining {
			if r == ':' || r == ' ' || r == '\t' {
				keyEnd = i
				break
			}
		}

		key = remaining[:keyEnd]

		if colonIdx != -1 && colonIdx == keyEnd {
			// Has ':' immediately after key
			remaining = remaining[colonIdx+1:]

			// Parse the value
			remaining = strings.TrimSpace(remaining)
			if len(remaining) > 0 && (remaining[0] == '"' || remaining[0] == '`' || remaining[0] == '\'') {
				// Quoted value - use strconv.Unquote
				// Find the end of the quoted string
				quote := remaining[0]
				endIdx := 1

				if quote == '`' {
					// Raw string literal
					for endIdx < len(remaining) {
						if remaining[endIdx] == '`' {
							endIdx++
							break
						}
						endIdx++
					}
				} else {
					// Regular quoted string
					escaped := false
					for endIdx < len(remaining) {
						if escaped {
							escaped = false
							endIdx++
							continue
						}
						if remaining[endIdx] == '\\' {
							escaped = true
							endIdx++
							continue
						}
						if remaining[endIdx] == quote {
							endIdx++
							break
						}
						endIdx++
					}
				}

				quotedStr := remaining[:endIdx]
				if unquoted, err := strconv.Unquote(quotedStr); err == nil {
					value = unquoted
					remaining = remaining[endIdx:]
				} else {
					// If unquote fails, treat as regular unquoted value
					if spaceIdx := strings.IndexAny(remaining, " \t"); spaceIdx != -1 {
						value = remaining[:spaceIdx]
						remaining = remaining[spaceIdx+1:]
					} else {
						value = remaining
						remaining = ""
					}
				}
			} else {
				// Unquoted value - take until next whitespace
				if spaceIdx := strings.IndexAny(remaining, " \t"); spaceIdx != -1 {
					value = remaining[:spaceIdx]
					remaining = remaining[spaceIdx+1:]
				} else {
					value = remaining
					remaining = ""
				}
			}
		} else {
			// No ':' or ':' is not immediately after key, it's a boolean flag
			value = ""
			if keyEnd < len(remaining) {
				remaining = remaining[keyEnd:]
			} else {
				remaining = ""
			}
		}

		result[key] = value
	}

	return result, nil
}

// Unmarshal converts ParsedDirective into a struct using reflection and struct tags
// Struct fields should have `directive:"name"` tags
// If the target implements UnmarshalDirective, that method is used instead
func (p ParsedDirective) Unmarshal(target interface{}) error {
	// Check if target implements UnmarshalDirective
	if unmarshaler, ok := target.(UnmarshalDirective); ok {
		return unmarshaler.UnmarshalDirective(p)
	}

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	v = v.Elem()
	t := v.Type()

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get the directive tag
		tag := field.Tag.Get("directive")
		if tag == "" || tag == "-" {
			continue
		}

		// Check if directive exists
		value, exists := p[tag]
		if !exists {
			continue
		}

		// Set the field value based on its type
		switch fieldValue.Kind() {
		case reflect.Bool:
			// For bool fields, presence means true (empty value is true)
			fieldValue.SetBool(true)
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if value != "" {
				if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
					fieldValue.SetInt(intVal)
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if value != "" {
				if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
					fieldValue.SetUint(uintVal)
				}
			}
		case reflect.Float32, reflect.Float64:
			if value != "" {
				if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					fieldValue.SetFloat(floatVal)
				}
			}
		}
	}

	return nil
}

// Legacy functions for compatibility

func EnumerateCommentGroup(comments *ast.CommentGroup) iter.Seq[string] {
	return func(yield func(string) bool) {
		if comments == nil || len(comments.List) == 0 {
			return
		}
		for _, c := range comments.List {
			if !yield(c.Text) {
				return
			}
		}
	}
}


func ParseFieldDirectiveCommentDst[T any](
	prefix string,
	comments dst.FieldDecorations,
	parser func(s []string) (T, error),
) (T, bool, error) {
	lines := extractDirectiveLines(
		hiter.Concat(
			slices.Values(afterLastEmptyLine(comments.Start)),
			slices.Values(clip1(comments.End)),
		),
		prefix,
		true,
	)
	if len(lines) == 0 {
		return *new(T), false, nil
	}
	t, err := parser(lines)
	return t, true, err
}

// Helper functions

func extractDirectiveLines(seq iter.Seq[string], directiveMarker string, allowNonDirective bool) []string {
	var stripped []string
	for comment := range seq {
		text := stripMarker(comment)
		if allowNonDirective {
			text = strings.TrimSpace(text)
		}
		var ok bool
		text, ok = strings.CutPrefix(text, directiveMarker)
		if !ok {
			if len(stripped) > 0 {
				break
			} else {
				continue
			}
		}
		stripped = append(stripped, text)
		// Only take the first directive line
		break
	}
	return stripped
}

func afterLastEmptyLine(lines []string) []string {
	var idx int
	for i, s := range slices.Backward(lines) {
		if s == "\n" { // needs strictly to be empty. dst doesn't handle lines with only white spaces correctly.
			if i > 0 && strings.HasPrefix(lines[i-1], "/*") {
				continue
			}
			idx = i + 1
			break
		}
	}
	return lines[idx:]
}

func clip1(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	return lines[:1]
}

func stripMarker(text string) string {
	if len(text) < 2 {
		return text
	}
	switch text[1] {
	case '/':
		return text[2:]
	case '*':
		return text[2 : len(text)-2]
	}
	return text
}

