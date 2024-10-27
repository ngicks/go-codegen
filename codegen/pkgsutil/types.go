package pkgsutil

import (
	"go/types"
	"iter"

	"github.com/ngicks/go-iterator-helper/hiter"
)

func EnumerateTypeParams(n *types.Named) iter.Seq[types.Type] {
	return func(yield func(types.Type) bool) {
		args := n.TypeArgs()
		for _, arg := range hiter.IndexAccessible(args, hiter.Range(0, args.Len())) {
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
