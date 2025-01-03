package matcher

import (
	"go/types"

	"github.com/ngicks/go-iterator-helper/hiter"
)

// IsNoCopy returns true given type ty has Lock() method,
// or the ty contains direct (not indirect by pointer, map, slice, channel) dependency to no-lock object.
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

func Name(ty types.Type) (pkgPath string, name string) {
	x, ok := ty.(interface{ Obj() *types.TypeName })
	if ok {
		if pkg := x.Obj().Pkg(); pkg != nil {
			pkgPath = pkg.Path()
		}
		name = x.Obj().Name()
		return
	}
	name = ty.String()
	return
}

func asInterface(ty types.Type) *types.Interface {
	i, _ := ty.(*types.Interface)
	return i
}

func asNamed(ty types.Type) *types.Named {
	n, _ := types.Unalias(ty).(*types.Named)
	return n
}

func as[T types.Type](ty types.Type) T {
	a, _ := types.Unalias(ty).(T)
	return a
}

func asPointer(ty types.Type) types.Type {
	switch x := ty.(type) {
	default:
		return ty
	case *types.Named:
		_, isInterface := types.Unalias(ty).Underlying().(*types.Interface)
		if isInterface {
			return ty
		}
		return types.NewPointer(x)
	}
}

func unwrapPointer(ty types.Type) types.Type {
	switch x := ty.(type) {
	default:
		return ty
	case *types.Pointer:
		return x.Elem()
	}
}

func findMethod(ty types.Type, methodName string) *types.Selection {
	ms := types.NewMethodSet(asPointer(ty))
	_, sel, _ := hiter.FindFunc2(
		func(_ int, sel *types.Selection) bool { return sel.Obj().Name() == methodName },
		hiter.AtterAll(ms),
	)
	return sel
}

// checks input sel is signature which takes no args and returns single value
func noArgSingleValue(sel *types.Selection) types.Type {
	if sel == nil {
		return nil
	}

	sig, ok := sel.Obj().Type().Underlying().(*types.Signature)
	if !ok {
		return nil
	}

	if sig.Params().Len() != 0 {
		return nil
	}

	tup := sig.Results()
	if tup.Len() != 1 {
		return nil
	}

	return tup.At(0).Type()
}

func identicalParametrized(i, j types.Type) bool {
	if types.Identical(i, j) {
		return true
	}

	iAlias, _ := i.(*types.Alias)
	jAlias, _ := j.(*types.Alias)

	iNamed, _ := i.(*types.Named)
	jNamed, _ := j.(*types.Named)

	return identicalOrigin(iAlias, jAlias, isNil) || identicalOrigin(iNamed, jNamed, isNil)
}

func isNil[T any](t *T) bool {
	return t == nil
}

func identicalOrigin[T interface{ Origin() U }, U *types.Alias | *types.Named](i, j T, isNil func(T) bool) bool {
	if isNil(i) || isNil(j) {
		return false
	}
	return types.Identical(types.Type(i.Origin()), types.Type(j.Origin()))
}
