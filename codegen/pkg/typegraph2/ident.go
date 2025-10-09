package typegraph2

import (
	"go/types"

	"github.com/ngicks/go-codegen/codegen/pkg/imports"
)

type Ident struct {
	PkgPath  string
	TypeName string
}

func NewIdentFromTypesObject(obj types.Object) Ident {
	var pkgsPath string
	if obj.Pkg() != nil {
		pkgsPath = obj.Pkg().Path()
	}
	return Ident{
		pkgsPath,
		obj.Name(),
	}
}

func NewIdentFromImportsTargetType(i imports.TargetType) Ident {
	return Ident{
		PkgPath:  i.ImportPath,
		TypeName: i.Name,
	}
}

func (t Ident) TargetType() imports.TargetType {
	return imports.TargetType{ImportPath: t.PkgPath, Name: t.TypeName}
}
