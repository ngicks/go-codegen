package astutil

import (
	"go/build/constraint"
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
