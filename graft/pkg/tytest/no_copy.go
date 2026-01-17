package tytest

import "go/types"

// IsNoCopy returns true given type ty has Lock() method,
// or ty contains direct (not indirected by pointer, map, slice, channel) dependency to no-lock object.
func IsNoCopy(ty types.Type) bool {
	sel := findMethod(ty, "Lock")
	if sel != nil {
		sig, ok := sel.Obj().Type().(*types.Signature)
		if !ok {
			return false
		}
		results := sig.Results()
		if sel.Obj().Name() == "Lock" && results.Len() == 0 {
			return true
		}
	}
	ty2 := types.Unalias(unwrapPointer(ty).Underlying())
	switch x := ty2.(type) {
	case *types.Named:
		return IsNoCopy(ty2)
	case *types.Struct:
		for i := range x.NumFields() {
			f := x.Field(i)
			if n := asNamed(f.Type()); n != nil {
				if asInterface(f.Type().Underlying()) == nil &&
					IsNoCopy(f.Type()) {
					return true
				}
			}
		}
		return false
	case *types.Array:
		n := asNamed(x.Elem())
		a := as[*types.Array](x.Elem())

		if n != nil || a != nil {
			if asInterface(x.Elem().Underlying()) == nil && IsNoCopy(x.Elem()) {
				return true
			}
		}
	}
	return false
}
