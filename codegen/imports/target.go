package imports

import (
	"path"
	"slices"
	"strings"
)

type TargetImports map[string]TargetImport

func (ti TargetImports) Append(imports ...TargetImport) {
	for _, imp := range imports {
		i := ti[imp.Import.Path]
		i.Import = imp.Import
		i.Ident = imp.Ident
		i.Types = append(i.Types, imp.Types...)
		slices.Sort(i.Types)
		i.Types = slices.Compact(i.Types)
		ti[imp.Import.Path] = i
	}
}

type TargetType struct {
	ImportPath string
	Name       string
}

type TargetImport struct {
	Import Import
	// may be empty
	Ident string
	Types []string
}

type Import struct {
	Path string
	Name string
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
