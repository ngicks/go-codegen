package pkgsutil

import (
	"go/types"
	"iter"
)

func EnumerateFields(n *types.Struct) iter.Seq2[int, *types.Var] {
	return func(yield func(int, *types.Var) bool) {
		for i := range n.NumFields() {
			if !yield(i, n.Field(i)) {
				return
			}
		}
	}
}
