package matcher

import (
	"go/types"
)

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
	named, _ := noArgSingleValue(sel).(*types.Named)
	if named == nil {
		return false
	}
	return named.Obj().Pkg() == nil && named.Obj().Name() == "error"
}

type ClonerMethod struct {
	// Method name
	Name string
}

func (method ClonerMethod) IsImplementor(ty types.Type) bool {
	sel := findMethod(ty, method.Name)
	ret := noArgSingleValue(sel)
	if ret == nil {
		return false
	}

	// receiver type is allowed to be pointer but
	// returned value must not be pointer.
	var unwrapped types.Type = ty
	switch x := ty.(type) {
	default:
		return false // is this even possible?
	case *types.Pointer:
		unwrapped = x.Elem()
	case *types.Alias, *types.Named, *types.Interface:
	}

	return types.Identical(ret, unwrapped)
}
