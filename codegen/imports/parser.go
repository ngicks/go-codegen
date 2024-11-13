package imports

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"maps"
	"slices"
	"strconv"

	"github.com/dave/dst"
	"github.com/dave/dst/dstutil"
	"github.com/ngicks/go-iterator-helper/hiter"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
	"golang.org/x/tools/go/packages"
)

type ImportParser struct {
	extra        TargetImports
	dependencies TargetImports
}

func NewParser(pkg *types.Package) *ImportParser {
	parser := &ImportParser{
		dependencies: make(TargetImports),
	}
	for p := range enumeratePkg(pkg) {
		parser.dependencies.Append(TargetImport{
			Import: Import{p.Path(), p.Name()},
			Types:  typesFromPkg(p),
		})
	}
	return parser
}

func NewParserPackages(pkgs []*packages.Package) *ImportParser {
	parser := &ImportParser{
		dependencies: make(TargetImports),
	}
	for p := range enumeratePkgs(pkgs) {
		parser.dependencies.Append(TargetImport{
			Import: Import{p.Path(), p.Name()},
			Types:  typesFromPkg(p),
		})
	}
	return parser
}

func typesFromPkg(pkg *types.Package) []string {
	return slices.Collect(
		xiter.Map(
			func(obj types.Object) string {
				return obj.Name()
			},
			xiter.Filter(
				func(obj types.Object) bool {
					_, ok1 := obj.Type().(*types.Named)
					_, ok2 := obj.Type().(*types.Alias)
					return ok1 || ok2
				},
				xiter.Map(
					func(s string) types.Object {
						return pkg.Scope().Lookup(s)
					},
					slices.Values(pkg.Scope().Names()),
				),
			),
		),
	)
}

func enumeratePkg(pkg *types.Package) iter.Seq[*types.Package] {
	return func(yield func(*types.Package) bool) {
		if !yield(pkg) {
			return
		}
		for _, i := range pkg.Imports() {
			for pkg := range enumeratePkg(i) {
				if !yield(pkg) {
					return
				}
			}
		}
	}
}

func enumeratePkgs(pkgs []*packages.Package) iter.Seq[*types.Package] {
	return func(yield func(*types.Package) bool) {
		packages.Visit(
			pkgs,
			func(p *packages.Package) bool {
				return yield(p.Types)
			},
			nil,
		)
	}
}

func (p *ImportParser) AppendExtra(imports ...TargetImport) {
	if p.extra == nil {
		p.extra = make(TargetImports)
	}
	p.extra.Append(imports...)
}

type ImportMap struct {
	// import ident to TargetImport
	ident   map[string]TargetImport
	missing map[string]TargetImport
	// package path to imports
	extra map[string]TargetImport
	// loaded from type info.
	dependencies map[string]TargetImport
}

func (p *ImportParser) Parse(importSpecs []*ast.ImportSpec) (ImportMap, error) {
	im := ImportMap{
		ident:        make(map[string]TargetImport),
		extra:        p.extra,
		dependencies: p.dependencies,
	}

	pkgPaths := make(map[string]bool)
	for _, is := range importSpecs {
		pkgPath := unquoteImportSpecPath(is)

		pkgPaths[pkgPath] = true

		var ident string
		if is.Name != nil {
			ident = is.Name.Name
		}
		if ident == "." || ident == "_" {
			continue
		}

		var ti TargetImport
		var ok bool
		ti, ok = im.extra[pkgPath]
		if !ok {
			ti, ok = im.dependencies[pkgPath]
			if !ok {
				return im, fmt.Errorf("unknown package path: %q", pkgPath)
			}
		}
		if ident == "" {
			ident = ti.Ident
			if ident == "" {
				if ti.Import.Name != "" {
					ident = ti.Import.Name
				} else {
					// import missing name
					// only possible case is caller-added extra imports
					// warn?
					ident = importPathToIdent(ti.Import.Path)
				}
				if importPathToIdent(ti.Import.Path) != ti.Import.Name {
					// cases like bubbletea; its package name is tea,
					// which confuses some tools like this one.
					ti.Ident = ident
				}
			}
		}
		// falling back is not expected to occur at this moment.
		// However user might have decided extra imports' ident name
		// and it's possible that the name overlapped to file's import.
		addFallingBack(im.ident, ident, ti)
	}

	hiter.ForEach2(
		func(pkgPath string, _ TargetImport) {
			_, _, _ = im.getIdent(pkgPath, "")
		},
		xiter.Filter2(
			func(s string, _ TargetImport) bool {
				return !pkgPaths[s]
			},
			maps.All(im.extra),
		),
	)

	return im, nil
}

// strips " or ` from import spec path.
func unquoteImportSpecPath(is *ast.ImportSpec) string {
	return unquoteBasicLitString(is.Path.Value)
}

// strips " or ` from basic lit string.
func unquoteBasicLitString(s string) string {
	if len(s) == 0 {
		// impossible. just avoiding panic. or should we panic?
		return s
	}
	if s[0] == '"' {
		pkgPath, err := strconv.Unquote(s)
		if err != nil {
			panic(fmt.Errorf("malformed import: %w", err))
		}
		return pkgPath
	} else {
		return s[1 : len(s)-1]
	}
}

func addFallingBack(m map[string]TargetImport, ident string, ti TargetImport) TargetImport {
	_, ok := m[ident]
	if !ok {
		m[ident] = ti
		return ti
	}
	for i := 1; ; i++ {
		_, ok := m[ident]
		if !ok {
			break
		}
		ident = ident + "_" + strconv.FormatInt(int64(i), 10)
		ti.Ident = ident
		continue
	}
	m[ident] = ti
	return ti
}

func (im *ImportMap) getIdent(pkgPath, name string) (string, TargetImport, bool) {
	for k, v := range im.ident {
		if v.Import.Path == pkgPath {
			if name != "" && !slices.Contains(v.Types, name) {
				return "", TargetImport{}, false
			}
			return k, v, true
		}
	}
	// fallback
	for _, m := range []map[string]TargetImport{im.extra, im.dependencies} {
		for _, v := range m {
			if v.Import.Path == pkgPath {
				if name != "" && !slices.Contains(v.Types, name) {
					return "", TargetImport{}, false
				}
				ident := firstNonEmpty(v.Ident, v.Import.Name, importPathToIdent(v.Import.Path))
				ti := addFallingBack(
					im.ident,
					ident,
					v,
				)
				ident = firstNonEmpty(ti.Ident, ident)
				im.recordMissing(ident, ti)
				return ident, ti, true
			}
		}
	}
	return "", TargetImport{}, false
}

func (im *ImportMap) recordMissing(ident string, ti TargetImport) {
	if im.missing == nil {
		im.missing = make(map[string]TargetImport)
	}
	if ti.Import.Name != ident {
		ti.Ident = ident
	}
	im.missing[ident] = ti
}

func firstNonEmpty[T comparable](ts ...T) T {
	var zero T
	for _, t := range ts {
		if t != zero {
			return t
		}
	}
	return zero
}

func (im ImportMap) AstExpr(ty TargetType) *ast.SelectorExpr {
	ident, _, ok := im.getIdent(ty.ImportPath, "") // whatever
	if !ok {
		return nil
	}

	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: ident,
		},
		Sel: &ast.Ident{
			Name: ty.Name,
		},
	}
}

func (im ImportMap) DstExpr(ty TargetType) *dst.SelectorExpr {
	ident, _, ok := im.getIdent(ty.ImportPath, "") // whatever
	if !ok {
		return nil
	}
	return &dst.SelectorExpr{
		X: &dst.Ident{
			Name: ident,
		},
		Sel: &dst.Ident{
			Name: ty.Name,
		},
	}
}

func (im *ImportMap) Ident(path string) (string, bool) {
	ident, _, ok := im.getIdent(path, "")
	return ident, ok
}

func (im ImportMap) MissingImports() iter.Seq2[string, string] {
	sorted := slices.SortedFunc(
		hiter.ToKeyValue(
			xiter.Map2(
				func(_ string, ti TargetImport) (string, string) {
					// it's ok that Ident is empty.
					if ti.Ident == "" && ti.Import.Name != importPathToIdent(ti.Import.Path) {
						return ti.Import.Name, ti.Import.Path
					}
					return ti.Ident, ti.Import.Path
				},
				maps.All(im.missing),
			),
		),
		func(i, j hiter.KeyValue[string, string]) int {
			return cmp.Compare(i.V, j.V)
		},
	)
	return hiter.Values2(sorted)
}

// AddMissingImports adds missing imports from imports to df,
// both [*dst.File.Imports] and tge first import decl in [*dst.File.Decls].
func (im ImportMap) AddMissingImports(df *dst.File) {
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
				definedImports := hiter.ReduceGroup(
					func(accum []string, next string) []string {
						// I suspect there are only 1-3 import decl for same package.
						// mostly only 1.
						// so keep it as slice, it should be slightly faster than map.
						return append(accum, next)
					},
					[]string(nil),
					hiter.Divide(
						func(is *dst.ImportSpec) (packagePath string, ident string) {
							if is.Name != nil {
								ident = is.Name.Name
							}
							return unquoteBasicLitString(is.Path.Value), ident
						},
						slices.Values(df.Imports),
					),
				)
				for ident, path := range im.MissingImports() {
					if slices.Contains(definedImports[path], ident) {
						continue
					}
					spec := &dst.ImportSpec{
						Name: dst.NewIdent(ident),
						Path: &dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(path)},
					}
					if ident == "" {
						spec.Name = nil
					}
					df.Imports = append(df.Imports, spec)
					x.Specs = append(x.Specs, spec)
				}
				// panic with bail-kind value then recover?
				replaced = true
				return false
			}
		},
		nil,
	)
}

func (im ImportMap) Qualifier() types.Qualifier {
	return func(p *types.Package) string {
		qual, _ := im.Ident(p.Path())
		return qual
	}
}
