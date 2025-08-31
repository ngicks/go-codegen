package directive

import (
	"go/ast"
	"go/types"
)

func ExcludeIgnoredGenDecl(genDecl *ast.GenDecl) (bool, error) {
	direction, _, err := ParseDirectiveComment(genDecl.Doc)
	if err != nil {
		return false, err
	}
	return !direction.MustIgnore(), nil
}

func ExcludeIgnoredTypeSpec(ts *ast.TypeSpec, _ types.Object) (bool, error) {
	direction, _, err := ParseDirectiveComment(ts.Doc)
	if err != nil {
		return false, err
	}
	return !direction.MustIgnore(), nil
}