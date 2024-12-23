package cloneruntime

import (
	"container/list"
	"container/ring"
	"reflect"
	"testing"
)

func TestCloneContainerList(t *testing.T) {
	l := list.New()

	l.PushBack(5)
	l.PushBack(7)
	l.PushBack(3)

	cloned := CloneContainerList(l)

	cloned.PushBack(9)

	var collected []int

	for ele := cloned.Front(); ele != nil; ele = ele.Next() {
		collected = append(collected, ele.Value.(int))
	}

	if !reflect.DeepEqual([]int{5, 7, 3, 9}, collected) {
		t.Fatal("wrong clone")
	}
	if l.Len() != 3 {
		t.Fatal("wrong clone")
	}
}

func TestCloneContainerRing(t *testing.T) {
	r := ring.New(3)

	r.Value = 5
	r = r.Next()
	r.Value = 7
	r = r.Next()
	r.Value = 3
	r = r.Next()

	cloned := CloneContainerRing(r)

	additive := ring.New(1)
	additive.Value = 9
	cloned.Prev().Link(additive)

	var collected []int
	cloned.Do(func(a any) {
		collected = append(collected, a.(int))
	})

	if !reflect.DeepEqual([]int{5, 7, 3, 9}, collected) {
		t.Fatal("wrong clone")
	}
	var leng int
	r.Do(func(a any) {
		leng++
	})
	if leng != 3 {
		t.Fatal("wrong clone")
	}
}
