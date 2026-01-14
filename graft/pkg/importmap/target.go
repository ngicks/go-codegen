package importmap

import (
	"go/types"
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
	PackagePath string
	TypeName    string
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
		// pkg is nil for built in types, e.g. error
		pkgPath = pkg.Path()
	}
	return t.PackagePath == pkgPath && t.TypeName == named.Obj().Name()
}
