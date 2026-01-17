package tytest

import "go/types"

// ErrorMethod describes a method that takes no argument and returns a single error value.
// Method name must be as Name.
type ErrorMethod struct {
	// Method name.
	Name string
}

// IsImplementor checks if ty implements a method named as [ErrorMethod.Name] that take no argument and returns an error.
func (method ErrorMethod) IsImplementor(ty *types.Named) bool {
	return isValidatorImplementor(ty, method.Name)
}

func isValidatorImplementor(ty *types.Named, methodName string) bool {
	sel := findMethod(ty, methodName)
	return IsError(noArgSingleValue(sel))
}

func IsError(ty types.Type) bool {
	named, _ := ty.(*types.Named)
	if named == nil {
		return false
	}
	if named.Obj() == nil {
		return false
	}
	return named.Obj().Pkg() == nil && named.Obj().Name() == "error"
}
