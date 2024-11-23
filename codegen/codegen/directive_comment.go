package codegen

import (
	"fmt"
	"go/ast"
	"iter"
	"slices"
	"strings"

	"github.com/dave/dst"
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

func ParseDirectiveComment(comments *ast.CommentGroup) (Direction, bool, error) {
	return parseDirective(EnumerateCommentGroup(comments))
}

func ParseDirectiveCommentDst(comments dst.NodeDecs) (Direction, bool, error) {
	var idx int
	for i, s := range slices.Backward(comments.Start) {
		if s == "\n" {
			if strings.HasPrefix(comments.Start[i-1], "/*") {
				continue
			}
			idx = i + 1
			break
		}
	}
	return parseDirective(slices.Values(comments.Start[idx:]))
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
