package tree

import "iter"

type Tree[T any] struct {
	node *node[T]
	//cloner:copyptr
	comparer func(i, j T) int
}

func New[T any](comparer func(i, j T) int) *Tree[T] {
	return &Tree[T]{
		comparer: comparer,
	}
}

func (t *Tree[T]) Push(ts ...T) {
	for _, val := range ts {
		if t.node == nil {
			t.node = &node[T]{
				data: val,
			}
		} else {
			t.node.push(val, t.comparer)
		}
	}
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

func (n *node[T]) push(t T, comparer func(i, j T) int) {
	if comparer(t, n.data) <= 0 {
		if n.l == nil {
			n.l = &node[T]{
				data: t,
			}
		} else {
			n.l.push(t, comparer)
		}
	} else {
		if n.r == nil {
			n.r = &node[T]{
				data: t,
			}
		} else {
			n.r.push(t, comparer)
		}
	}
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
