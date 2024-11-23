package codegen

import (
	"go/ast"
	"iter"
)

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
