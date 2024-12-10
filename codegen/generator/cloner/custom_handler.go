package cloner

import (
	"go/types"
	"slices"

	"github.com/ngicks/go-codegen/codegen/imports"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
)

type CustomHandlers []CustomHandler

func (h CustomHandlers) Match(ty types.Type) int {
	for i, handler := range h {
		if handler.Matcher(ty) {
			return i
		}
	}
	return -1
}

func (h CustomHandlers) Imports() []imports.TargetImport {
	return slices.Collect(
		hiter.Flatten(
			xiter.Map(
				func(h CustomHandler) []imports.TargetImport {
					return h.Imports
				},
				slices.Values(h),
			),
		),
	)
}

type CustomHandler struct {
	Matcher func(types.Type) bool
	Imports []imports.TargetImport
	Expr    func(imports.ImportMap) func(s string) string
}

var builtinCustomHandlers = [...]CustomHandler{
	{
		// slices.Clone on basic slice types.
		Matcher: func(t types.Type) bool {
			s, ok := t.(*types.Slice)
			if !ok {
				return false
			}
			// Only for basic types (or known clone-by-assign safe types)
			// Technically we can safely invoke slices.Clone on slice of any clone-by-assign types
			// but this Matcher signature can't determine whether they are a hand-written Clone implementor or just a generation target thus an implementor.
			// Generation targets aren't implementors on the first run, but are them on the second or later run.
			// TODO: expand interface and detect implementor or matched type?
			_, ok = s.Elem().(*types.Basic)
			// Basic includes unsafe pointer and uintptr but I assume placing them is fully intentional and safe to copy.
			return ok || isKnownCloneByAssign(s.Elem())
		},
		Imports: []imports.TargetImport{
			{
				Import: imports.Import{Path: "slices", Name: "slices"},
			},
		},
		Expr: func(im imports.ImportMap) func(s string) string {
			return func(s string) string {
				ident, _ := im.Ident("slices")
				return ident + ".Clone(" + s + ")"
			}
		},
	},
	{
		// calls maps.Clone on basic map type.
		// maps.Clone uses the internal mechanism so some (possibly future) performance improvement is expected.
		Matcher: func(t types.Type) bool {
			s, ok := t.(*types.Map)
			if !ok {
				return false
			}
			// Same as above. Only for basic types or known clone-by-assign.
			_, ok = s.Elem().(*types.Basic)
			return ok || isKnownCloneByAssign(s.Elem())
		},
		Imports: []imports.TargetImport{
			{
				Import: imports.Import{Path: "maps", Name: "maps"},
			},
		},
		Expr: func(im imports.ImportMap) func(s string) string {
			return func(s string) string {
				ident, _ := im.Ident("maps")
				return ident + ".Clone(" + s + ")"
			}
		},
	},
	{
		// call cloneruntime.Time on time.Time
		// that clones time but strips monotonic timer.
		Matcher: func(t types.Type) bool {
			return imports.TargetType{ImportPath: "time", Name: "Time"}.Is(t)
		},
		Imports: []imports.TargetImport{
			{
				Import: imports.Import{
					Path: "github.com/ngicks/go-codegen/pkg/cloneruntime",
					Name: "cloneruntime",
				},
			},
		},
		Expr: func(im imports.ImportMap) func(s string) string {
			return func(s string) string {
				ident, _ := im.Ident("github.com/ngicks/go-codegen/pkg/cloneruntime")
				return ident + ".Time(" + s + ")"
			}
		},
	},
	{
		// just do nothing for known clone-by-assign types.
		Matcher: func(t types.Type) bool {
			return isKnownCloneByAssign(t)
		},
		Expr: func(im imports.ImportMap) func(s string) string {
			return func(s string) string {
				return s
			}
		},
	},
}

var knownCloneByAssign = map[imports.TargetType]struct{}{
	{ImportPath: "unique", Name: "Handle"}: {},
}

func isKnownCloneByAssign(ty types.Type) bool {
	named, ok := ty.(*types.Named)
	if !ok {
		return false
	}
	var pkgPath string
	if pkg := named.Obj().Pkg(); pkg != nil {
		pkgPath = pkg.Path()
	}
	_, ok = knownCloneByAssign[imports.TargetType{ImportPath: pkgPath, Name: named.Obj().Name()}]
	return ok
}
