package tree

import "iter"

type Tree[T any] struct {
	node *node[T]
}

func (t *Tree[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		t.node.walk(yield)
	}
}

type node[T any] struct {
	l, r *node[T]
	data T
}

func (n *node[T]) walk(yield func(T) bool) bool {
	if n.l != nil && !n.l.walk(yield) {
		return false
	}
	if !yield(n.data) {
		return false
	}
	if n.r != nil && !n.r.walk(yield) {
		return false
	}
	return true
}
