package matcher

import "go/types"

func IsCloneByAssign(ty types.Type, allowNamed bool) bool {
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
		if !allowNamed {
			return false
		}
		return IsCloneByAssign(ty.Underlying(), allowNamed)
	case *types.Struct:
		for i := range x.NumFields() {
			ty := x.Field(i).Type()
			if !IsCloneByAssign(ty, allowNamed) {
				return false
			}
		}
		return true
	case *types.Array:
		return IsCloneByAssign(x.Elem(), allowNamed)
	}
}
