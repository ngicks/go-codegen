package typematcher

import "go/types"

func IsCloneByAssign(ty types.Type, stepNext func(*types.Named) bool) bool {
	// Don't worry about type recursion
	// we are not traversing on pointers
	// which is needed for type recursion.
	switch x := ty.(type) {
	default:
		return false
	case *types.Basic:
		// both uintptr and unsafe.Pointer are just thought as a mere numeric value.
		return true
	case *types.Named:
		if !stepNext(x) {
			return false
		}
		return IsCloneByAssign(ty.Underlying(), stepNext)
	case *types.Struct:
		for i := range x.NumFields() {
			ty := x.Field(i).Type()
			if !IsCloneByAssign(ty, stepNext) {
				return false
			}
		}
		return true
	case *types.Array:
		return IsCloneByAssign(x.Elem(), stepNext)
	}
}
