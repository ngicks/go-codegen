package undgen

import (
	"fmt"
	"go/ast"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/dave/dst"
	"github.com/ngicks/go-iterator-helper/x/exp/xiter"
)

type TargetType struct {
	ImportPath string
	Name       string
}

// importDecls maps idents (PackageNames) to TargetImport
type importDecls struct {
	identToImport map[string]TargetImport
	fallback      map[string]TargetImport
}

// parseImports relates ident (PackageName) to TargetImport.
func parseImports(importSpecs []*ast.ImportSpec, imports []TargetImport) importDecls {
	// pre-process input
	imports = slices.Collect(
		xiter.Map(
			func(t TargetImport) TargetImport {
				t.Types = slices.Clone(t.Types)
				return t
			},
			slices.Values(imports),
		),
	)

	id := importDecls{
		identToImport: map[string]TargetImport{},
		fallback:      map[string]TargetImport{},
	}

	for _, is := range importSpecs {
		pkgPath := unquoteImportSpecPath(is)
		idx := slices.IndexFunc(imports, func(i TargetImport) bool { return pkgPath == i.ImportPath })
		if idx < 0 {
			continue
		}
		targetImport := imports[idx]
		imports = slices.Delete(imports, idx, idx+1)
		id.identToImport[identImportSpec(is)] = targetImport
	}

	importSpecNames := slices.Collect(xiter.Map(identImportSpec, slices.Values(importSpecs)))
	// now `imports` is residue of target imports.
	// Maybe we swill need to refer or add these.
	// We just store them as fallback.
	// Name ident that does not overlap to existing import specs and store elements of `imports` associating with it.
	for _, i := range imports {
		name := importPathToIdent(i.ImportPath)
		for i := 1; ; i++ {
			_, ok := id.fallback[name]
			if !ok && !slices.Contains(importSpecNames, name) {
				break
			}
			name = name + "_" + strconv.FormatInt(int64(i), 10)
			continue
		}
		id.fallback[name] = i
	}
	return id
}

// strips " or ` from import spec path.
func unquoteImportSpecPath(is *ast.ImportSpec) string {
	if is.Path.Value[0] == '"' {
		pkgPath, err := strconv.Unquote(is.Path.Value)
		if err != nil {
			panic(fmt.Errorf("malformed import: %w", err))
		}
		return pkgPath
	} else {
		return is.Path.Value[1 : len(is.Path.Value)-1]
	}
}

// returns ident accessing import spec.
func identImportSpec(is *ast.ImportSpec) string {
	if is.Name != nil {
		return is.Name.Name
	}
	return importPathToIdent(unquoteImportSpecPath(is))
}

// converts import path to ident accessing import spec.
// If path is suffixed with major version (`v`%d), then base name of path prefix is returned.
func importPathToIdent(pkgPath string) string {
	pkgBase := path.Base(pkgPath)
	if strings.HasPrefix(pkgBase, "v") && len(strings.TrimFunc(pkgBase[1:], isAsciiNum)) == 0 {
		pkgBase = path.Base(path.Dir(pkgPath))
	}
	return pkgBase
}

func isAsciiNum(r rune) bool {
	return '0' <= r && r <= '9'
}

func (id importDecls) matchIdentToImport(pkgPath, name string) (string, TargetImport) {
	for k, v := range id.identToImport {
		if v.ImportPath == pkgPath && slices.Contains(v.Types, name) {
			return k, v
		}
	}
	return "", TargetImport{}
}

func (id importDecls) matchFallback(pkgPath, name string) (string, TargetImport) {
	for k, v := range id.fallback {
		if v.ImportPath == pkgPath && slices.Contains(v.Types, name) {
			return k, v
		}
	}
	return "", TargetImport{}
}

func (id importDecls) DstExpr(pkgPath, name string) *dst.SelectorExpr {
	var (
		ident string
	)
	ident, _ = id.matchIdentToImport(pkgPath, name)
	if ident == "" {
		ident, _ = id.matchFallback(pkgPath, name)
	}

	if ident == "" {
		return nil
	}

	return &dst.SelectorExpr{
		X: &dst.Ident{
			Name: ident,
		},
		Sel: &dst.Ident{
			Name: name,
		},
	}
}

func (id importDecls) Len() int {
	return len(id.identToImport)
}

func (id importDecls) HasSelector(left, right string) bool {
	ti, ok := id.identToImport[left]
	return ok && slices.Contains(ti.Types, right)
}

func (i importDecls) MatchTy(path string, name string) (TargetType, bool) {
	for _, v := range i.identToImport {
		if v.ImportPath == path && slices.Contains(v.Types, name) {
			return TargetType{v.ImportPath, name}, true
		}
	}
	return TargetType{}, false
}
