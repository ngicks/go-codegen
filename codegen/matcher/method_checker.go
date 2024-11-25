package matcher

import (
	"go/types"

	"github.com/ngicks/go-iterator-helper/hiter"
)

type CyclicConversionMethods struct {
	From    bool
	Reverse string
	Convert string
}

func (mset CyclicConversionMethods) IsImplementor(ty *types.Named) bool {
	_, ok := isCyclicConversionMethodsImplementor(ty, mset)
	return ok
}

func (mset CyclicConversionMethods) ConvertedType(ty *types.Named) (*types.Named, bool) {
	return isCyclicConversionMethodsImplementor(ty, mset)
}

// isCyclicConversionMethodsImplementor checks if ty can be converted to a type, then converted back from the type to ty
// though methods described in conversionMethod.
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

	ms := types.NewMethodSet(asPointer(ty))
	for _, sel := range hiter.AtterAll(ms) {
		if sel.Obj().Name() == toMethod {
			sig, ok := sel.Obj().Type().Underlying().(*types.Signature)
			if !ok {
				return nil, false
			}
			tup := sig.Results()
			if tup.Len() != 1 {
				return nil, false
			}
			v := tup.At(0)

			toType, ok := v.Type().(*types.Named)
			if !ok {
				return nil, false
			}

			ms := types.NewMethodSet(asPointer(toType))
			for _, sel := range hiter.AtterAll(ms) {
				if sel.Obj().Name() != revMethod {
					continue
				}

				sig, ok := sel.Obj().Type().Underlying().(*types.Signature)
				if !ok {
					return toType, false
				}
				tup := sig.Results()
				if tup.Len() != 1 {
					return toType, false
				}
				v := tup.At(0)

				supposeToBeFromType, ok := v.Type().(*types.Named)
				if !ok {
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
		}
	}
	return nil, false
}

type ErrorMethod struct {
	Name string
}

func (method ErrorMethod) IsImplementor(ty *types.Named) bool {
	return isValidatorImplementor(ty, method.Name)
}

func isValidatorImplementor(ty *types.Named, methodName string) bool {
	ms := types.NewMethodSet(types.NewPointer(ty))
	for i := range ms.Len() {
		sel := ms.At(i)
		if sel.Obj().Name() == methodName {
			sig, ok := sel.Obj().Type().Underlying().(*types.Signature)
			if !ok {
				return false
			}
			tup := sig.Results()
			if tup.Len() != 1 {
				return false
			}
			v := tup.At(0)

			named, ok := v.Type().(*types.Named)
			if !ok {
				return false
			}
			return named.Obj().Pkg() == nil && named.Obj().Name() == "error"
		}
	}
	return false
}
