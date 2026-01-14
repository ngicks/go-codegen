// package cyclic defines cyclicly-referenced type like tree.
package cyclic

import (
	"iter"

	"github.com/ngicks/go-codegen/graft/pkg/typegraph/testdata/target"
)

type Tree struct {
	l, r   *Tree
	value  any
	Marker faketarget.FakeTarget
}

func (t *Tree) Value() any {
	return t.value
}

func (t *Tree) Iter() iter.Seq[any] {
	// just here for realism and shut up the linter.
	return func(yield func(any) bool) {
		if t.l != nil {
			for v := range t.l.Iter() {
				if !yield(v) {
					return
				}
			}
		}
		if !yield(t.value) {
			return
		}
		if t.r != nil {
			for v := range t.r.Iter() {
				if !yield(v) {
					return
				}
			}
		}
	}
}

type CyclicEmbedded struct {
	Marker target.FakeTarget
	recursion1
}

type recursion1 struct {
	recursion2
}

type recursion2 struct {
	*CyclicEmbedded
}
