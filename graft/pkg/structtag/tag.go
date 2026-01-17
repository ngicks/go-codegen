// Package structtag provides utilities for parsing and manipulating Go struct tags.
//
// Struct tags follow the format:
//
//	tagname:"name,opt1,opt2,opt3:value"
//
// Where:
//   - tagname is the tag key (e.g., "json", "xml", "db")
//   - name is the field name or primary value
//   - options can be simple identifiers (e.g., "omitempty") or key-value pairs (e.g., "format:RFC3339")
//
// The name portion can be single-quoted to escape special characters:
//
//	json:"'\xde\xad\xbe\xef',omitempty"
//
// This package extends the standard library's reflect.StructTag with mutation capabilities,
// allowing programmatic addition, deletion, and retrieval of tag options.
package structtag

// This file uses modified Go programming language standard library.
// So keep it credited.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Modified parts are governed by a license that is described in ../LICENSE.

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// ErrUnpairedKey is returned when a tag key has no paired quoted value.
	// For example, `json:foo` (missing quotes) or `json:"` (unterminated quote).
	ErrUnpairedKey = errors.New("unpaired key")
	// ErrNotFound is returned when the requested tag name or option does not exist.
	ErrNotFound = errors.New("not found")
	// ErrSyntax is returned when the input string has invalid syntax for unescaping.
	ErrSyntax = errors.New("syntax error")
)

// Tag represents a single key-value pair from a struct tag.
// For example, the struct tag `json:"foo,omitempty"` produces a Tag
// with Key="json" and Value="foo,omitempty".
type Tag struct {
	Key   string
	Value string
}

// String returns the tag formatted as key:"value" with the value properly quoted.
func (t Tag) String() string {
	return t.Key + ":" + strconv.Quote(t.Value)
}

// Tags is a slice of Tag representing all tags on a struct field.
// It provides methods for querying and modifying individual tag options.
type Tags []Tag

// ParseStructTag parses a reflect.StructTag into a slice of Tag.
// It returns ErrUnpairedKey if any tag key lacks a properly quoted value.
func ParseStructTag(tag reflect.StructTag) (Tags, error) {
	var out []Tag

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			return nil, fmt.Errorf("%w: input has no paired value, rest = %s", ErrUnpairedKey, string(tag))
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return nil, fmt.Errorf("%w: name = %s has no paired value, rest = %s", ErrUnpairedKey, name, string(tag))
		}
		quotedValue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(quotedValue)
		if err != nil {
			return nil, err
		}
		out = append(out, Tag{Key: name, Value: value})
	}

	return out, nil
}

// Of returns [reflect.StructTag] consits of tags.
func Of(tags []Tag) reflect.StructTag {
	var buf strings.Builder
	for _, tag := range tags {
		buf.Write([]byte(tag.String()))
		buf.WriteByte(' ')
	}

	out := buf.String()
	if len(out) > 0 {
		out = out[:len(out)-1]
	}
	return reflect.StructTag(out)
}

func (t Tags) mapTag(tagName string, option string, fn func(v string, n, m int, err error) (string, error)) (Tags, error) {
	// TODO: defer this clone until really needed.
	tt := slices.Clone(t)

	for i, tag := range tt {
		if tag.Key != tagName {
			continue
		}
		n, m, _, err := getRange(tag.Value, option)
		tag.Value, err = fn(tag.Value, n, m, err)
		if err != nil {
			return tt, err
		}
		tt[i] = tag
		return tt, nil
	}

	return tt, ErrNotFound
}

// Get retrieves the value of an option from the specified tag.
//
// Assuming tags are formatted like `json:"name,opt1,opt2:value"`,
// specifying option retrieves that option (e.g. opt1 fopt opt1, opt2:value for opt2)
// If options is empty, it returns first portion of tag (e.g. name of json:"name,opt".)
//
// escaped != unescaped if a corresponding part of tag value is escaped in t.
//
// Returns [ErrNotFound] if the tag or option does not exist.
func (t Tags) Get(tagName string, option string) (escaped, unescaped string, err error) {
	for _, tag := range t {
		if tag.Key != tagName {
			continue
		}
		n, m, unescaped, err := getRange(tag.Value, option)
		if err != nil {
			return "", "", err
		}
		if unescaped == "" {
			unescaped = tag.Value[n:m]
		}
		return tag.Value[n:m], unescaped, nil
	}
	return "", "", ErrNotFound
}

// Delete returns a new Tags whose part corresponding to tagName and option is removed.
// If option is empty, it removes the name portion.
//
// Returns [ErrNotFound] if the tag or option does not exist.
func (t Tags) Delete(tagName string, option string) (Tags, error) {
	return t.mapTag(tagName, option, func(v string, n, m int, err error) (string, error) {
		if err != nil {
			return v, err
		}

		if option != "" && len(v) > m && v[m] == ',' {
			m++
		}

		v = v[:n] + v[m:]
		// maybe an empty string and a single comma is left behind. cut it.
		v, _ = strings.CutSuffix(v, ",")

		return v, nil
	})
}

// Add returns a new Tags with tagName:",option:value" added.
//
// Either option or value can be empty.
// If option is empty, value is added as name portion (i.e. tagName:"value")
// If value is empty, option is added as an option without value. (i.e. tagName:",option")
//
// If the option already exists, the original Tags is returned unchanged.
func (t Tags) Add(tagName string, option, value string) (Tags, error) {
	tt, err := t.mapTag(tagName, option, func(v string, n, m int, err error) (string, error) {
		if err == nil {
			return v, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return v, err
		}

		// option does not exist in the tagName:"" section. Add an option.

		if option == "" {
			return value + v, nil
		}

		v = v + "," + option
		if value != "" {
			v += ":" + value
		}

		return v, nil
	})
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return tt, err
		}

		// tagName:"" itself does not exist.
		var newValue string
		if option == "" {
			newValue = value
		} else {
			newValue = "," + option
			if value != "" {
				newValue += ":" + value
			}
		}

		tt = append(tt, Tag{
			Key:   tagName,
			Value: newValue,
		})
	}

	return tt, nil
}

// StructTag converts Tags to a reflect.StructTag.
//
// This is equivalent to calling Of(t).
func (t Tags) StructTag() reflect.StructTag {
	return Of(t)
}

func getRange(tag string, targetOption string) (n, m int, unescaped string, err error) {
	// first, skip name.
	if len(tag) > 0 && !strings.HasPrefix(tag, ",") {
		unescaped, n, err := readName(tag)
		if err != nil {
			return -1, -1, "", err
		}
		if targetOption == "" {
			return 0, n, unescaped, nil
		}
		m = n
		tag = tag[n:]
	}

	for len(tag) > 0 {
		n = m + 1
		if tag[0] != ',' {
			return -1, -1, "", ErrNotFound
		} else {
			tag = tag[1:]
			if len(tag) == 0 {
				return -1, -1, "", ErrNotFound
			}
		}

		var (
			opt string
			mm  int
		)
		opt, mm, err = readTagOption(tag)
		if err != nil {
			return -1, -1, "", err
		}

		m = n + mm

		tag = tag[mm:]
		if len(tag) > 0 && tag[0] == ':' {
			tag = tag[len(":"):]
			_, mm, err = readTagOption(tag)
			if err != nil {
				return
			}
			tag = tag[mm:]
			m += len(":") + mm
		}

		if opt == targetOption {
			return n, m, "", nil
		}
	}

	return -1, -1, "", ErrNotFound
}

// AddTagOption returns a new StructTag which has option added for tag.
// It assumes tag options are formatted as `tag:"name,opt,opt"` style.
// The name is allowed to be quoted by single quotation marks.
func AddTagOption(t reflect.StructTag, tag string, option string) (reflect.StructTag, error) {
	tags, err := ParseStructTag(t)
	if err != nil {
		return "", err
	}

	hasTag := false
	for i := 0; i < len(tags); i++ {
		if tags[i].Key != tag {
			continue
		}

		hasTag = true

		hasValue := false

		value := tags[i].Value
		// first, skip name.
		if len(value) > 0 && !strings.HasPrefix(value, ",") {
			_, n, err := readName(value)
			if err != nil {
				return "", err
			}
			value = value[n:]
		}

		for len(value) > 0 {
			if value[0] != ',' {
				return "", fmt.Errorf("malformed option, %s", tags[i].Value)
			} else {
				value = value[1:]
				if len(value) == 0 {
					return "", fmt.Errorf("malformed option, %s", tags[i].Value)
				}
			}

			opt, n, err := readTagOption(value)
			if err != nil {
				return "", err
			}

			value = value[n:]
			if len(value) > 0 && value[0] == ':' {
				if strings.HasPrefix(option, opt+":") {
					hasValue = true
					break
				}
				value = value[len(":"):]
				_, n, err := readTagOption(value)
				if err != nil {
					return "", err
				}
				value = value[n:]
			}

			if option == opt {
				hasValue = true
				break
			}
		}

		if !hasValue {
			if !strings.HasPrefix(option, ",") {
				tags[i].Value += ","
			}
			tags[i].Value += option
		}
		break
	}

	if !hasTag {
		tags = append(tags, Tag{Key: tag, Value: option})
	}

	return Of(tags), nil
}

func readName(value string) (unescaped string, n int, err error) {
	n = len(value) - len(strings.TrimLeftFunc(value, func(r rune) bool {
		return !strings.ContainsRune(",\\'\"`", r) // reserve comma, backslash, and quotes
	}))
	if n == 0 {
		unescaped, n, err = readTagOption(value)
	}
	return
}

func readTagOption(s string) (opt string, n int, err error) {
	if len(s) == 0 {
		return "", 0, io.ErrUnexpectedEOF
	}

	switch r, _ := utf8.DecodeRuneInString(s); {
	case r == '_' || unicode.IsLetter(r): // Go ident
		n = len(s) - len(strings.TrimLeftFunc(s, func(r rune) bool {
			return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
		}))
		return s[:n], n, nil
	case r == '\'': // escaped
		return unescape(s)
	default:
		return "", 0, fmt.Errorf("invalid character: %s", s)
	}
}

func unescape(s string) (unescaped string, n int, err error) {
	i := 0
	if s[0] == '\'' {
		i = 1
	}

	escaping := false
	escaped := []byte{'"'}
	for i < len(s) {
		r, rn := utf8.DecodeRuneInString(s[i:])
		switch {
		case escaping:
			if r == '\'' {
				escaped = escaped[:len(escaped)-1]
			}
			escaping = false
		case r == '\\':
			escaping = true
		case r == '"':
			escaped = append(escaped, '\\')
		case r == '\'':
			escaped = append(escaped, '"')
			i += 1
			out, err := strconv.Unquote(string(escaped))
			if err != nil {
				return "", 0, fmt.Errorf("invalid escaped string: string must be escaped by single quotes, input = %s", s)
			}
			return out, i, nil
		}
		escaped = append(escaped, s[i:][:rn]...)
		i += rn
	}
	return "", 0, fmt.Errorf("invalid escaped string: single-quoted string missing terminating single-quote: %s", s)
}

