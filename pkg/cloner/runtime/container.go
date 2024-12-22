package cloneruntime

import (
	"container/list"
	"container/ring"
)

func CloneContainerList(l *list.List) *list.List {
	if l == nil {
		return nil
	}
	new := list.New()
	for ele := l.Front(); ele != nil; ele = ele.Next() {
		new.PushBack(ele.Value)
	}
	return new
}

func CloneContainerRing(r *ring.Ring) *ring.Ring {
	if r == nil {
		return nil
	}
	new := ring.New(r.Len())
	r.Do(func(a any) {
		new.Value = a
		new = new.Next()
	})
	return new
}
