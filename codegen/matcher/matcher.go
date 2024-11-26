package matcher

import (
	"go/types"

	"github.com/ngicks/go-iterator-helper/hiter"
)

func IsNoCopy(ty types.Type) bool {
	sel := findMethod(ty, "Lock")
	if sel == nil {
		return false
	}
	sig, ok := sel.Obj().Type().(*types.Signature)
	if !ok {
		return false
	}
	results := sig.Results()
	return sel.Obj().Name() == "Lock" && results.Len() == 0
}

func asPointer(ty types.Type) types.Type {
	ty = types.Unalias(ty)
	switch x := ty.(type) {
	default:
		return ty
	case *types.Named:
		if _, isInterface := x.Underlying().(*types.Interface); isInterface {
			return ty
		}
		return types.NewPointer(x)
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
