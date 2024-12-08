package codegen

import (
	"go/ast"
	"go/types"
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

func ExcludeIgnoredGenDecl(genDecl *ast.GenDecl) (bool, error) {
	direction, _, err := ParseDirectiveComment(genDecl.Doc)
	if err != nil {
		return false, err
	}
	return !direction.MustIgnore(), nil
}

func ExcludeIgnoredTypeSpec(ts *ast.TypeSpec, _ types.Object) (bool, error) {
	direction, _, err := ParseDirectiveComment(ts.Doc)
	if err != nil {
		return false, err
	}
	return !direction.MustIgnore(), nil
}
