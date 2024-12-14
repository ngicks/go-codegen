package cloner

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"

	"github.com/ngicks/go-codegen/codegen/codegen"
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
	Expr    func(CustomHandlerExprData) (expr func(s string) (expr string), isFunc bool)
}

type CustomHandlerExprData struct {
	ImportMap imports.ImportMap
	AstExpr   ast.Expr
	Ty        types.Type
}

var builtinCustomHandlers = [...]CustomHandler{
	{
		// copy on basic slice types.
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

			// Basic includes unsafe pointer and uintptr but I assume placing them is fully intentional and safe to copy.
			return isBasicOrKnownCloneByAssign(s.Elem())
		},
		Imports: []imports.TargetImport{
			{
				Import: imports.Import{Path: "slices", Name: "slices"},
			},
		},
		Expr: func(data CustomHandlerExprData) (expr func(s string) (expr string), isFunc bool) {
			return func(s string) string {
				return fmt.Sprintf( // You can't use slices.Clone here since it does not copy cap.
					`func(src %[1]s) %[1]s {
						if src == nil {
							return nil
						}
						dst := make(%[1]s, len(src), cap(src))
						copy(dst, src)
						return dst
					}`, codegen.PrintAstExprPanicking(data.AstExpr))
			}, true
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
			return isBasicOrKnownCloneByAssign(s.Elem())
		},
		Imports: []imports.TargetImport{
			{
				Import: imports.Import{Path: "maps", Name: "maps"},
			},
		},
		Expr: func(data CustomHandlerExprData) (expr func(s string) (expr string), isFunc bool) {
			return func(s string) string {
				ident, _ := data.ImportMap.Ident("maps")
				return ident + ".Clone"
			}, true
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
					Path: "time",
					Name: "time",
				},
			},
		},
		Expr: func(data CustomHandlerExprData) (expr func(s string) (expr string), isFunc bool) {
			return func(s string) string {
				ident, _ := data.ImportMap.Ident("time")
				tok := "t"
				if tok == ident {
					tok = "tt"
				}
				return fmt.Sprintf(
					`func(%[1]s %[2]s.Time) %[2]s.Time {
						return %[2]s.Date(
							%[1]s.Year(),
							%[1]s.Month(),
							%[1]s.Day(),
							%[1]s.Hour(),
							%[1]s.Minute(),
							%[1]s.Second(),
							%[1]s.Nanosecond(),
							%[1]s.Location(),
						)
					}`, tok, ident)
			}, true
		},
	},
	{
		// just do nothing for array of bare known clone-by-assign types.
		Matcher: func(t types.Type) bool {
			if a, ok := t.(*types.Array); ok {
				t = a.Elem()
			}
			return isBasicOrKnownCloneByAssign(t)
		},
		Expr: func(data CustomHandlerExprData) (expr func(s string) (expr string), isFunc bool) {
			return func(s string) string {
				return s
			}, false
		},
	},
}

func isBasicOrKnownCloneByAssign(ty types.Type) bool {
	if _, ok := ty.(*types.Basic); ok {
		return true
	}
	return isKnownCloneByAssign(ty)
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
