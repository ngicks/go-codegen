package astutil

import (
	"go/ast"
	"go/build/constraint"
	"go/types"
	"iter"
	"slices"

	"github.com/dave/dst"
	"github.com/ngicks/go-iterator-helper/hiter"
)

func TrimPackageComment(f *dst.File) {
	// we only support Go 1.21+ since the package "maps" is first introduced in that version.
	// The "// +build" is no longer supported after Go 1.18
	// but we still leave comments as long as it is easy to implement.
	f.Decs.Start = slices.AppendSeq(
		dst.Decorations{},
		hiter.Filter(
			func(s string) bool { return constraint.IsGoBuild(s) || constraint.IsPlusBuild(s) },
			slices.Values(f.Decs.Start),
		),
	)
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