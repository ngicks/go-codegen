package importmap

import (
	"go/types"
	"path"
	"strings"
)

// Specs is set of import spec.
type Specs map[string]Spec

func (is Specs) Insert(specs ...Spec) {
	for _, s := range specs {
		i := is[s.Package.Path]
		i.Package = s.Package
		is[s.Package.Path] = i
	}
}

// Spec represents import spec.
type Spec struct {
	Package Package
}

func specFromTypesPackage(pkg *types.Package) Spec {
	return Spec{
		Package: packageFromTypesPackage(pkg),
	}
}

type Package struct {
	// Import path to the package.
	Path string
	// Pacakge name, i.e. math for "math/v2"
	Name string
}

func packageFromTypesPackage(pkg *types.Package) Package {
	return Package{
		pkg.Path(),
		pkg.Name(),
	}
}

type QualifiedType struct {
	ImportPath string
	TypeName   string
}

func (t QualifiedType) Is(ty types.Type) bool {
	named, ok := ty.(*types.Named)
	if !ok {
		return false
	}
	if named.Obj() == nil {
		return false
	}
	pkg := named.Obj().Pkg()
	var pkgPath string
	if pkg != nil {
		pkgPath = pkg.Path()
	}
	return t.ImportPath == pkgPath && t.TypeName == named.Obj().Name()
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
