// Package tytest provides type testers for identifying types with specific characteristics
// that are too complex to be expressed as an interface (e.g. NoCopy types, interface implementers).
package tytest

import (
	"go/types"

	"github.com/ngicks/go-iterator-helper/hiter"
)

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
	sel, _ := hiter.FindFunc(
		func(sel *types.Selection) bool { return sel.Obj().Name() == methodName },
		ms.Methods(),
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
