package directive

import (
	"fmt"
	"go/ast"
	"iter"
	"slices"
	"strings"

	"github.com/dave/dst"
	"github.com/ngicks/go-iterator-helper/hiter"
)

const (
	DirectivePrefix = "codegen:"
)

const (
	DirectiveCommentIgnore    = "ignore"
	DirectiveCommentGenerated = "generated"
)

type Direction struct {
	ignore    bool
	generated bool
}

func (d Direction) MustIgnore() bool {
	return d.ignore || d.generated
}

func (d Direction) IsGenerated() bool {
	return d.generated
}

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

func ParseDirectiveComment(comments *ast.CommentGroup) (Direction, bool, error) {
	return parseDirective(EnumerateCommentGroup(comments))
}

func ParseDirectiveCommentDst(comments dst.NodeDecs) (Direction, bool, error) {
	return parseDirective(slices.Values(afterLastEmptyLine(comments.Start)))
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

func ParseFieldDirectiveCommentDst[T any](
	prefix string,
	comments dst.FieldDecorations,
	parser func(s []string) (T, error),
) (T, bool, error) {
	lines := directiveComments(
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

func parseDirective(seq iter.Seq[string]) (Direction, bool, error) {
	direction := directiveComments(seq, DirectivePrefix, true)

	var dir Direction
	if len(direction) == 0 {
		return dir, false, nil
	}

	switch direction[0] {
	default:
		return dir, true, fmt.Errorf("unknown: %v", direction)
	case DirectiveCommentIgnore:
		dir.ignore = true
	case DirectiveCommentGenerated:
		dir.generated = true
	}

	return dir, true, nil
}

func directiveComments(seq iter.Seq[string], directiveMarker string, allowNonDirective bool) []string {
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
	}
	return stripped
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