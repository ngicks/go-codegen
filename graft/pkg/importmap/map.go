// Package importmap defines parser and modifier for import specs in Go's source code.
package importmap

import (
	"cmp"
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	pathpkg "path"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
)

type NamedSpec struct {
	// Renamed ident for import spec, i.e. import ident "fully-qualified/package/path"
	Ident string
	Spec  Spec
}

func (ns NamedSpec) Name() (string, bool) {
	if ns.Ident != "" {
		return ns.Ident, true
	}
	if ns.Spec.Package.Name != "" {
		return ns.Spec.Package.Name, false
	}
	return inferPackageName(ns.Spec.Package.Path), true
}

func inferPackageName(path string) string {
	pkgBase := pathpkg.Base(path)
	if strings.HasPrefix(pkgBase, "v") && len(strings.TrimFunc(pkgBase[1:], isAsciiNum)) == 0 {
		pkgBase = pathpkg.Base(pathpkg.Dir(path))
	}
	return pkgBase
}

func compareNamedSpec(l, r NamedSpec) int {
	if c := cmp.Compare(l.Spec.Package.Path, r.Spec.Package.Path); c != 0 {
		return c
	}
	return cmp.Compare(l.Ident, r.Ident)
}

func cleanNamedSpec(specs []NamedSpec) []NamedSpec {
	slices.SortFunc(
		specs,
		compareNamedSpec,
	)
	return slices.CompactFunc(
		specs,
		func(l, r NamedSpec) bool {
			return l.Spec.Package == r.Spec.Package && l.Ident == r.Ident
		},
	)
}

func searchNamedSpecs(specs []NamedSpec, tgt NamedSpec) (NamedSpec, bool) {
	i, ok := slices.BinarySearchFunc(
		specs,
		tgt,
		compareNamedSpec,
	)
	if !ok {
		return NamedSpec{}, false
	}
	return specs[i], true
}

func searchNamedSpecsByPath(specs []NamedSpec, pkgPath string) (NamedSpec, bool) {
	i, ok := slices.BinarySearchFunc(
		specs,
		pkgPath,
		func(ns NamedSpec, path string) int {
			return cmp.Compare(ns.Spec.Package.Path, path)
		},
	)
	if !ok {
		return NamedSpec{}, false
	}
	return specs[i], true
}

type Map struct {
	// set of idents with which one can access imported packages.
	names map[string]struct{}
	// import specs existing in parsed file.
	ident []NamedSpec
	// specs already been added in p's added field or requested at run-time but was missing.
	missing []NamedSpec
	p       *Parser
}

// getIdent returns the identifier used to access a package by its path.
// If the package is not found in the existing imports, it falls back to
// dependencies and records it as missing.
func (m *Map) getIdent(pkgPath string) (string, NamedSpec, bool) {
	if ns, ok := searchNamedSpecsByPath(m.ident, pkgPath); ok {
		name, _ := ns.Name()
		return name, ns, true
	}

	if ns, ok := searchNamedSpecsByPath(m.missing, pkgPath); ok {
		name, _ := ns.Name()
		return name, ns, true
	}

	spec, ok := m.p.dependencies[pkgPath]
	if !ok {
		return "", NamedSpec{}, false
	}

	ns := NamedSpec{Spec: spec}
	added := m.appendFallingBackClean(ns)
	name, _ := added.Name()
	return name, added, true
}

// Ident returns the identifier used to access a package by its path.
// If the package was not in the original imports, it is recorded as missing.
func (m *Map) Ident(pkgPath string) (string, bool) {
	ident, _, ok := m.getIdent(pkgPath)
	return ident, ok
}

// AstExpr returns an *ast.SelectorExpr for accessing a type from the given qualified type.
// Returns nil if the package is not found.
func (m *Map) AstExpr(ty QualifiedType) *ast.SelectorExpr {
	ident, ok := m.Ident(ty.ImportPath)
	if !ok {
		return nil
	}

	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: ident,
		},
		Sel: &ast.Ident{
			Name: ty.TypeName,
		},
	}
}

// DstExpr returns a *dst.SelectorExpr for accessing a type from the given qualified type.
// Returns nil if the package is not found.
func (m *Map) DstExpr(ty QualifiedType) *dst.SelectorExpr {
	ident, ok := m.Ident(ty.ImportPath)
	if !ok {
		return nil
	}

	return &dst.SelectorExpr{
		X: &dst.Ident{
			Name: ident,
		},
		Sel: &dst.Ident{
			Name: ty.TypeName,
		},
	}
}

// MissingImports returns an iterator over (ident, path) pairs of imports
// that need to be added to the file. The results are sorted by path.
// The ident is non-empty when an explicit import name is needed
// (either because it was explicitly set, or because the package name
// differs from the path-inferred name).
func (m *Map) MissingImports() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, ns := range m.missing {
			var ident string
			if ns.Ident != "" {
				// Explicit ident was set
				ident = ns.Ident
			} else if ns.Spec.Package.Name != "" && ns.Spec.Package.Name != inferPackageName(ns.Spec.Package.Path) {
				// Package name differs from path-inferred name, need explicit ident
				ident = ns.Spec.Package.Name
			}
			if !yield(ident, ns.Spec.Package.Path) {
				return
			}
		}
	}
}

// AddMissingImports adds missing imports to the dst.File,
// updating both File.Imports and the first import declaration in File.Decls.
func (m *Map) AddMissingImports(df *dst.File) {
	definedImports := make(map[string][]string)
	for _, is := range df.Imports {
		var ident string
		if is.Name != nil {
			ident = is.Name.Name
		}
		path := unquoteBasicLitString(is.Path.Value)
		definedImports[path] = append(definedImports[path], ident)
	}

	var replaced bool
	dstutil.Apply(
		df,
		func(c *dstutil.Cursor) bool {
			if replaced {
				return false
			}
			node := c.Node()
			switch x := node.(type) {
			default:
				return true
			case *dst.GenDecl:
				if x.Tok != token.IMPORT {
					return false
				}
				for ident, path := range m.MissingImports() {
					if slices.Contains(definedImports[path], ident) {
						continue
					}
					spec := &dst.ImportSpec{
						Path: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(path)},
					}
					if ident != "" {
						spec.Name = dst.NewIdent(ident)
					}
					df.Imports = append(df.Imports, spec)
					x.Specs = append(x.Specs, spec)
				}
				replaced = true
				return false
			}
		},
		nil,
	)
}

// Qualifier returns a types.Qualifier that fully qualifies members of all
// packages other than currentPkgPath.
func (m *Map) Qualifier(currentPkgPath string) types.Qualifier {
	return func(p *types.Package) string {
		if currentPkgPath == p.Path() {
			return ""
		}
		qual, _ := m.Ident(p.Path())
		return qual
	}
}
