package tytest

import "go/types"

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
	unwrapped := unwrapPointer(ty)
	switch unwrapped.(type) {
	default:
		return false // is this even possible?
	case *types.Alias, *types.Named:
	}

	return types.Identical(ret, unwrapped)
}

// *types.Alias, *types.Named
type hasTypeParam interface {
	TypeParams() *types.TypeParamList
	TypeArgs() *types.TypeList
}

func (method ClonerMethod) IsFuncImplementor(ty types.Type) bool {
	// unwrap single pointer *T -> T then check type params and args.
	// The type may still be wrapped in pointer but double (or more) pointer type can not be a method receiver.
	// Thus it can be ignored anyway.
	parametrizedType, ok := unwrapPointer(ty).(hasTypeParam)
	if !ok {
		return false
	}

	if parametrizedType.TypeParams().Len() == 0 {
		return false
	}

	sel := findMethod(ty, method.Name+"Func")
	if sel == nil {
		return false
	}

	sig, ok := sel.Obj().Type().Underlying().(*types.Signature)
	if !ok {
		return false
	}

	if sig.Params().Len() != parametrizedType.TypeParams().Len() {
		return false
	}

	for i := range parametrizedType.TypeParams().Len() {
		sig, ok := sig.Params().At(i).Type().(*types.Signature)
		if !ok {
			return false
		}

		if sig.Params().Len() != 1 {
			return false
		}

		if !identicalParamOrArg(sig.Params().At(0).Type(), i, parametrizedType.TypeParams(), parametrizedType.TypeArgs()) {
			return false
		}

		if sig.Results().Len() != 1 {
			return false
		}

		if !identicalParamOrArg(sig.Results().At(0).Type(), i, parametrizedType.TypeParams(), parametrizedType.TypeArgs()) {
			return false
		}
	}

	tup := sig.Results()
	if tup.Len() != 1 {
		return false
	}

	return identicalParametrized(unwrapPointer(ty), tup.At(0).Type())
}

func identicalParamOrArg(sigTy types.Type, i int, params *types.TypeParamList, args *types.TypeList) bool {
	switch x := sigTy.(type) {
	case *types.TypeParam:
		return x.Index() == params.At(i).Index()
	default:
		return args.Len() > i && types.Identical(sigTy, args.At(i))
	}
}
