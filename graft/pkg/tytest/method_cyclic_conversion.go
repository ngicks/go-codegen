package tytest

import (
	"go/types"
)

// TODO: return diagnose error instead of simple boolean?
// The interfaces defined and checked here are relatively complex, at least more complex than the simple interface satisfaction check.
// Users might be confused why some specific types aren't an implementor.

// CyclicConversionMethods describes method that convert a type A to another type B through Convert,
// which can be converted back to A through Reverse.
// If From is false, an input type is assumed to be A, otherwise B.
type CyclicConversionMethods struct {
	From    bool
	Reverse string
	Convert string
}

// IsImplementor check if given type ty is implementor of mset.
// Methods can be implemented on pointer receiver.
func (mset CyclicConversionMethods) IsImplementor(ty *types.Named) bool {
	_, ok := isCyclicConversionMethodsImplementor(ty, mset)
	return ok
}

// ConvertedType returns the type converted through Convert or Reverse depending on From.
// The returned value is true only if ty is implementor of mset,
// in that case returned [*types.Named] is guaranteed to be non-nil.
func (mset CyclicConversionMethods) ConvertedType(ty *types.Named) (*types.Named, bool) {
	return isCyclicConversionMethodsImplementor(ty, mset)
}

// isCyclicConversionMethodsImplementor checks if ty can be converted to a type, then converted back from the type to ty
// through methods described in methods.
//
// Assuming fromPlain is false, ty is an implementor if ty (called type A hereafter)
// has the method which [CyclicConversionMethods.Convert] names
// where the returned value of the method is only one and type B,
// and also type B implements the method which [CyclicConversionMethods.Reverse] describes
// where the returned value of the method is only one and type A.
//
// If fromPlain is true isCyclicConversionMethodsImplementor works reversely (it checks assuming ty is type B.)
func isCyclicConversionMethodsImplementor(ty *types.Named, methods CyclicConversionMethods) (*types.Named, bool) {
	toMethod := methods.Convert
	revMethod := methods.Reverse
	if methods.From {
		toMethod, revMethod = revMethod, toMethod
	}

	sel := findMethod(ty, toMethod)
	toType, _ := noArgSingleValue(sel).(*types.Named)
	if toType == nil {
		return nil, false
	}

	sel = findMethod(toType, revMethod)
	supposeToBeFromType, _ := noArgSingleValue(sel).(*types.Named)
	if supposeToBeFromType == nil {
		return toType, false
	}

	if types.Identical(ty, supposeToBeFromType) {
		return toType, true
	}
	// they aren't identical. but is ty un-instantiated?
	// If yes then, check again with instantiated type
	if types.Identical(ty, supposeToBeFromType.Origin()) &&
		ty.TypeArgs().Len() == 0 &&
		supposeToBeFromType.TypeArgs().Len() > 0 {
		toType2, ok := isCyclicConversionMethodsImplementor(supposeToBeFromType, methods)
		if !ok {
			return toType, false
		}
		return toType, types.Identical(toType, toType2)
	}
	return toType, false
}
