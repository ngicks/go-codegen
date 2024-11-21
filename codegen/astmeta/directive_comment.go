package astmeta

import (
	"fmt"
	"go/ast"
	"strings"
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

func ParseComment(comments *ast.CommentGroup) (Direction, bool, error) {
	direction := directiveComments(comments, DirectivePrefix, true)

	var ud Direction
	if len(direction) == 0 {
		return ud, false, nil
	}

	switch direction[0] {
	default:
		return ud, true, fmt.Errorf("unknown: %v", direction)
	case DirectiveCommentIgnore:
		ud.ignore = true
	case DirectiveCommentGenerated:
		ud.generated = true
	}

	return ud, true, nil
}

func directiveComments(cg *ast.CommentGroup, directiveMarker string, allowNonDirective bool) []string {
	if cg == nil || cg.List == nil {
		return nil
	}
	var stripped []string
	for _, c := range cg.List {
		text := stripMarker(c.Text)
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
