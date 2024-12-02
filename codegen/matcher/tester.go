package matcher

import "go/types"

func IsCloneByAssign(ty types.Type) bool {
	// as you can see, no traversal on pointer
	// that's why you don't need to worry about
	// type recursion.
	switch x := ty.(type) {
	default:
		return false
	case *types.Basic:
		return x.Kind() != types.UnsafePointer
	case *types.Named:
		return IsCloneByAssign(ty.Underlying())
	case *types.Struct:
		for i := range x.NumFields() {
			ty := x.Field(i).Type()
			if !IsCloneByAssign(ty) {
				return false
			}
		}
		return true
	case *types.Array:
		return IsCloneByAssign(x.Elem())
	}
}
