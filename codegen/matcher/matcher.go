package matcher

import (
	"go/types"

	"github.com/ngicks/go-iterator-helper/hiter"
)

func IsNoCopy(ty types.Type) bool {
	mset := types.NewMethodSet(asPointer(ty))
	for _, sel := range hiter.AtterAll(mset) {
		sig, ok := sel.Obj().Type().(*types.Signature)
		if !ok {
			continue
		}
		results := sig.Results()
		if sel.Obj().Name() == "Lock" && results.Len() == 0 {
			return true
		}
	}
	return false
}

func asPointer(ty types.Type) types.Type {
	switch x := ty.(type) {
	default:
		return ty
	case *types.Named:
		if _, isInterface := x.Underlying().(*types.Interface); isInterface {
			return ty
		}
		return types.NewPointer(x)
	case *types.Alias:
		// alias rhs may still be aliased.
		return asPointer(x.Rhs())
	}
}

func IsCloner(ty types.Type) bool {
	return false

}
