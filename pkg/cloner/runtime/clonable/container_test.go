package clonable

import (
	"cmp"
	"reflect"
	"slices"
	"testing"
)

type indexedNum struct {
	index int
	value int
}

func TestSliceHeap_simple(t *testing.T) {
	h := SliceHeap[int]{
		Interface: SliceInterface[int]{
			Cmp: cmp.Compare[int],
		},
	}

	h.Push(5)
	h.Push(7)
	h.Push(3)

	cloned := h.CloneFunc(func(i int) int { return i })

	var popped []int
	for cloned.Interface.Len() > 0 {
		popped = append(popped, cloned.Pop())
	}

	expected := []int{3, 5, 7}
	if !reflect.DeepEqual(expected, popped) {
		t.Fatalf("wrong impl:\nexpected = %#v\nactual= %#v", expected, popped)
	}

	if h.Interface.Len() != 3 {
		t.Fatalf(
			"manipulation to cloned object propagates to original: expected len == 3, actual = %d",
			h.Interface.Len(),
		)
	}

	h.Interface.Cmp = func(i, j int) int {
		return -cmp.Compare(i, j)
	}
	h.Init()

	cloned2 := h.CloneFunc(func(i int) int { return i })

	var popped2 []int
	for cloned2.Interface.Len() > 0 {
		popped2 = append(popped2, cloned2.Pop())
	}

	expected = []int{7, 5, 3}
	if !reflect.DeepEqual(expected, popped2) {
		t.Fatalf("wrong impl:\nexpected = %#v\nactual= %#v", expected, popped2)
	}
}

func TestSliceHeap_hooked(t *testing.T) {
	pushArg := []indexedNum{}

	h := SliceHeap[indexedNum]{
		Interface: SliceInterface[indexedNum]{
			Cmp: func(i, j indexedNum) int {
				return cmp.Compare[int](i.value, j.value)
			},
			Hooks: SliceInterfaceHooks[indexedNum]{
				Pop: func(iface *SliceInterface[indexedNum], beingPopped *indexedNum) {
					t.Helper()
					if iface.Slice[iface.Len()-1] != *beingPopped {
						t.Fatalf(
							"pop is not being called right before value is popped: "+
								"slice = %#v, beingPopped = %d",
							iface.Slice, beingPopped,
						)
					}
					beingPopped.index = -1
				},
				Push: func(iface *SliceInterface[indexedNum], beingPushed *indexedNum) {
					pushArg = append(pushArg, *beingPushed)
					beingPushed.index = iface.Len()
				},
				Swap: func(iface *SliceInterface[indexedNum], i, j int) {
					iface.Slice[i].index = j
					iface.Slice[j].index = i
				},
			},
		},
	}

	h.Push(indexedNum{value: 5})
	h.Push(indexedNum{value: 7})
	h.Push(indexedNum{value: 3})

	t.Run("hook usage", func(t *testing.T) {
		expectedPushed := []indexedNum{{value: 5}, {value: 7}, {value: 3}}
		if !reflect.DeepEqual(expectedPushed, pushArg) {
			t.Fatalf("wrong push hook usage:\nexpected = %#v\nactual = %#v\n", expectedPushed, pushArg)
		}
	})

	t.Run("correct heap implementation", func(t *testing.T) {
		expected := []indexedNum{{0, 3}, {1, 7}, {2, 5}}
		if !reflect.DeepEqual(expected, h.Interface.Slice) {
			t.Fatalf("wrong clone:\nexpected = %#v\nactual = %#v\n", expected, h.Interface.Slice)
		}
	})

	t.Run("CloneFunc", func(t *testing.T) {
		cloned := h.CloneFunc(func(in indexedNum) indexedNum { return in })

		var popped []indexedNum
		for cloned.Interface.Len() > 0 {
			popped = append(popped, cloned.Pop())
		}

		expectedPopped := []indexedNum{{-1, 3}, {-1, 5}, {-1, 7}}
		if !reflect.DeepEqual(expectedPopped, popped) {
			t.Fatalf("wrong clone:\nexpected = %#v\nactual = %#v\n", expectedPopped, popped)
		}

		if h.Interface.Len() != 3 {
			t.Fatalf("expected len == 3 but is %d", h.Interface.Len())
		}

		cloned = h.CloneFunc(func(in indexedNum) indexedNum { return in })

		cloned.Interface.Swap(0, 1)
		expectedSwapped := []indexedNum{{0, 7}, {1, 3}, {2, 5}}
		if !reflect.DeepEqual(expectedSwapped, cloned.Interface.Slice) {
			t.Fatalf("wrong clone:\nexpected = %#v\nactual = %#v\n", expectedSwapped, cloned.Interface.Slice)
		}

		cloned.Init() // fix

		cloned.Push(indexedNum{-5, 9})

		expectedPushed := []indexedNum{{value: 5}, {value: 7}, {value: 3}, {index: -5, value: 9}}
		if !reflect.DeepEqual(expectedPushed, pushArg) {
			t.Fatalf("wrong hook clone:\nexpected = %#v\nactual = %#v\n", expectedPushed, pushArg)
		}
	})
}

func TestList(t *testing.T) {
	l := NewList[int]()

	for _, num := range []int{5, 7, 3} {
		l.PushBack(num)
	}

	collect := func(l List[int]) []int {
		return slices.Collect(l.Front().ValuesForward())
	}

	t.Run("CloneFunc", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i * 2 })
		cloned.PushBack(9)

		collected := collect(cloned)

		expected := []int{10, 14, 6, 9}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
		if l.Len() != 3 {
			t.Fatalf(
				"manipulation to cloned object propagates to the original."+
					"expected len == 3, but is %d",
				l.Len(),
			)
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		if l.Unwrap().Front().Value != 5 {
			t.Fatalf("wrong impl")
		}
		cloned := l.CloneFunc(func(i int) int { return i })
		if l.Unwrap() == cloned.Unwrap() {
			t.Fatalf("wrong impl")
		}
	})

	// The rest of lines tests all methods... should not be interesting to your eye.

	t.Run("Back", func(t *testing.T) {
		if l.Back().Get() != 3 {
			t.Fatalf("wrong impl")
		}
	})
	t.Run("Front", func(t *testing.T) {
		if l.Front().Get() != 5 {
			t.Fatalf("wrong impl")
		}
	})
	t.Run("Init", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })
		cloned.Init()
		if cloned.Len() != 0 {
			t.Fatal("wrong impl")
		}
	})
	t.Run("InsertAfter", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })

		cloned.InsertAfter(12, cloned.Front().Next())

		collected := collect(cloned)
		expected := []int{5, 7, 12, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("InsertBefore", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })

		cloned.InsertBefore(12, cloned.Front().Next())

		collected := collect(cloned)
		expected := []int{5, 12, 7, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("Len", func(t *testing.T) {
		if l.Len() != 3 {
			t.Fatalf("wrong impl")
		}
	})
	t.Run("MoveAfter", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })
		cloned.MoveAfter(cloned.Back(), cloned.Front())
		collected := collect(cloned)
		expected := []int{5, 3, 7}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("MoveBefore", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })
		cloned.MoveBefore(cloned.Back(), cloned.Front())
		collected := collect(cloned)
		expected := []int{3, 5, 7}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("MoveToBack", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })
		cloned.MoveToBack(cloned.Front().Next())
		collected := collect(cloned)
		expected := []int{5, 3, 7}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("MoveToFront", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })
		cloned.MoveToFront(cloned.Front().Next())
		collected := collect(cloned)
		expected := []int{7, 5, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("PushBack", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })
		cloned.PushBack(12)
		collected := collect(cloned)
		expected := []int{5, 7, 3, 12}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("PushBackList", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })

		additive := NewList[int]()
		additive.PushBack(12)
		additive.PushBack(1)

		cloned.PushBackList(additive)

		collected := collect(cloned)
		expected := []int{5, 7, 3, 12, 1}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("PushFront", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })
		cloned.PushFront(12)
		collected := collect(cloned)
		expected := []int{12, 5, 7, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("PushFrontList", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })

		additive := NewList[int]()
		additive.PushBack(12)
		additive.PushBack(1)

		cloned.PushFrontList(additive)

		collected := collect(cloned)
		expected := []int{12, 1, 5, 7, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("Remove", func(t *testing.T) {
		cloned := l.CloneFunc(func(i int) int { return i })

		cloned.Remove(cloned.Front().Next())

		collected := collect(cloned)
		expected := []int{5, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
}

func TestRing(t *testing.T) {
	r := NewRing[int](3)

	if r.Get() != 0 {
		t.Fatalf("wrong impl")
	}

	count := 0
	r.Do(func(i int) {
		t.Helper()
		count++
		if i != 0 {
			t.Fatalf("wrong initialization")
		}
	})

	if count != 3 {
		t.Fatalf("wrong len")
	}

	var i int
	for rr := range r.Forward() {
		i++
		rr.Set(i)
	}

	t.Run("iterators", func(t *testing.T) {
		var expected, collected []int
		collected = slices.Collect(r.ValuesForward())
		expected = []int{1, 2, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
		collected = slices.Collect(r.ValuesBackward())
		expected = []int{1, 3, 2}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("CloneFunc", func(t *testing.T) {
		cloned := r.CloneFunc(func(i int) int { return i * 2 })

		collected := slices.Collect(cloned.ValuesForward())
		expected := []int{2, 4, 6}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}

		collected = slices.Collect(r.ValuesForward())
		expected = []int{1, 2, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})

	t.Run("Do", func(t *testing.T) {
		collected := []int{}
		r.Do(func(i int) {
			collected = append(collected, i)
		})
		expected := []int{1, 2, 3}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("Len", func(t *testing.T) {
		if r.Len() != 3 {
			t.Fatal("wrong impl")
		}
	})
	t.Run("Link", func(t *testing.T) {
		cloned := r.CloneFunc(func(i int) int { return i })

		additive := NewRing[int](2)
		for rr := range additive.Forward() {
			rr.Set(9)
		}
		if cloned.Prev().Link(additive).Unwrap() != cloned.Unwrap() {
			t.Fatalf("wrong impl")
		}
		collected := slices.Collect(cloned.ValuesForward())
		expected := []int{1, 2, 3, 9, 9}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
	t.Run("Move", func(t *testing.T) {
		cloned := r.CloneFunc(func(i int) int { return i })
		if cloned.Move(3).Get() != 1 {
			t.Fatal("wrong impl")
		}
		if cloned.Move(-1).Get() != 3 {
			t.Fatal("wrong impl")
		}
	})
	t.Run("Next", func(t *testing.T) {
		rr := r.Next()
		if rr.Get() != 2 {
			t.Fatal("wrong impl")
		}

		rr = rr.Next()
		if rr.Get() != 3 {
			t.Fatal("wrong impl")
		}

		rr = rr.Next()
		if rr.Get() != 1 {
			t.Fatal("wrong impl")
		}

		if rr.Unwrap() != r.Unwrap() {
			t.Fatal("wrong impl")
		}
	})
	t.Run("Prev", func(t *testing.T) {
		rr := r.Prev()
		if rr.Get() != 3 {
			t.Fatal("wrong impl")
		}

		rr = rr.Prev()
		if rr.Get() != 2 {
			t.Fatal("wrong impl")
		}

		rr = rr.Prev()
		if rr.Get() != 1 {
			t.Fatal("wrong impl")
		}

		if rr.Unwrap() != r.Unwrap() {
			t.Fatal("wrong impl")
		}
	})
	t.Run("Unlink", func(t *testing.T) {
		cloned := r.CloneFunc(func(i int) int { return i })

		additive := NewRing[int](2)
		for rr := range additive.Forward() {
			rr.Set(9)
		}
		cloned.Prev().Link(additive)

		unlinked := cloned.Unlink(3)

		collected := slices.Collect(cloned.ValuesForward())
		expected := []int{1, 9}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}

		collected = slices.Collect(unlinked.ValuesForward())
		expected = []int{2, 3, 9}
		if !reflect.DeepEqual(expected, collected) {
			t.Fatalf("not equal:\nexpected = %#v\nactual = %#v", expected, collected)
		}
	})
}
