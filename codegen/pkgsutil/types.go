package pkgsutil

import (
	"go/types"
	"iter"

	"github.com/ngicks/go-iterator-helper/hiter"
)

func EnumerateTypeParams(n *types.Named) iter.Seq[types.Type] {
	return func(yield func(types.Type) bool) {
		args := n.TypeArgs()
		for _, arg := range hiter.AtterAll(args) {
			if !yield(arg) {
				return
			}
			named, ok := arg.Underlying().(*types.Named)
			if !ok {
				continue
			}
			for ty := range EnumerateTypeParams(named) {
				if !yield(ty) {
					return
				}
			}
		}
	}
}

func EnumerateFields(n *types.Struct) iter.Seq2[int, *types.Var] {
	return func(yield func(int, *types.Var) bool) {
		for i := range n.NumFields() {
			if !yield(i, n.Field(i)) {
				return
			}
		}
	}
}
