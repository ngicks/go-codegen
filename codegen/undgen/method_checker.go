package undgen

import (
	"go/types"

	"github.com/ngicks/go-iterator-helper/hiter"
)

type ConversionMethodsSet struct {
	FromPlain bool
	ToRaw     string
	ToPlain   string
}

func (mset ConversionMethodsSet) IsImplementor(ty *types.Named) bool {
	_, ok := isConversionMethodImplementor(ty, mset, mset.FromPlain)
	return ok
}

func (mset ConversionMethodsSet) ConvertedType(ty *types.Named) (*types.Named, bool) {
	return isConversionMethodImplementor(ty, mset, mset.FromPlain)
}

// isConversionMethodImplementor checks if ty can be converted to a type, then converted back from the type to ty
// though methods described in conversionMethod.
//
// Assuming fromPlain is false, ty is an implementor if ty (called type A hereafter)
// has the method which [ConversionMethodsSet.ToPlain] names
// where the returned value of the method is only one and type B,
// and also type B implements the method which [ConversionMethodsSet.ToRaw] describes
// where the returned value of the method is only one and type A.
//
// If fromPlain is true isConversionMethodImplementor works reversely (it checks assuming ty is type B.)
func isConversionMethodImplementor(ty *types.Named, conversionMethod ConversionMethodsSet, fromPlain bool) (*types.Named, bool) {
	toMethod := conversionMethod.ToPlain
	revMethod := conversionMethod.ToRaw
	if fromPlain {
		toMethod, revMethod = revMethod, toMethod
	}

	ms := types.NewMethodSet(types.NewPointer(ty))
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

			ms := types.NewMethodSet(types.NewPointer(toType))
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

				objStr1 := ty.Obj().String() // Assigning to a value just to inspect the string in the debugger.
				objStr2 := supposeToBeFromType.Obj().String()
				// simple pointer comparison should not suffice since
				// if types are instantiated, they can be same type but different pointer.
				// Am I correct? At least if I replace the line below with `return ty == named`
				// Test_isImplementor fails.
				return toType, objStr1 == objStr2
			}
		}
	}
	return nil, false
}

type ValidatorMethod struct {
	Name string
}

func (method ValidatorMethod) IsImplementor(ty *types.Named) bool {
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
