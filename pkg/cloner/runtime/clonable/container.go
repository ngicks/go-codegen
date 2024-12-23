package clonable

import (
	"container/heap"
	"container/list"
	"container/ring"
	"iter"
)

var _ heap.Interface = (*SliceInterface[any])(nil)

type SliceInterface[T any] struct {
	Cmp   func(i, j T) int
	Hooks SliceInterfaceHooks[T]
	Slice []T
}

type SliceInterfaceHooks[T any] struct {
	Pop  func(iface *SliceInterface[T], beingPopped *T)
	Push func(iface *SliceInterface[T], beingPushed *T)
	Swap func(iface *SliceInterface[T], i, j int)
}

func (s SliceInterface[T]) CloneFunc(cloneT func(T) T) SliceInterface[T] {
	var cloned []T
	if s.Slice != nil {
		cloned = make([]T, len(s.Slice), cap(s.Slice))
		for i, t := range s.Slice {
			cloned[i] = cloneT(t)
		}
	}
	return SliceInterface[T]{
		Cmp:   s.Cmp,
		Hooks: s.Hooks,
		Slice: cloned,
	}
}

func (s *SliceInterface[T]) Len() int {
	return len(s.Slice)
}

func (s *SliceInterface[T]) Less(i int, j int) bool {
	return s.Cmp(s.Slice[i], s.Slice[j]) < 0
}

func (s *SliceInterface[T]) Pop() any {
	popped := s.Slice[len(s.Slice)-1]

	if s.Hooks.Pop != nil {
		s.Hooks.Pop(s, &popped)
	}

	var zero T
	s.Slice[len(s.Slice)-1] = zero
	s.Slice = s.Slice[:len(s.Slice)-1]

	return popped
}

func (s *SliceInterface[T]) Push(x any) {
	t := x.(T)

	if s.Hooks.Push != nil {
		s.Hooks.Push(s, &t)
	}

	s.Slice = append(s.Slice, t)
}

func (s *SliceInterface[T]) Swap(i int, j int) {
	if s.Hooks.Swap != nil {
		s.Hooks.Swap(s, i, j)
	}
	s.Slice[i], s.Slice[j] = s.Slice[j], s.Slice[i]
}

type SliceHeap[T any] struct {
	Interface SliceInterface[T]
}

func (h SliceHeap[T]) CloneFunc(cloneT func(T) T) SliceHeap[T] {
	return SliceHeap[T]{Interface: h.Interface.CloneFunc(cloneT)}
}

func (h *SliceHeap[T]) Fix(i int)      { heap.Fix(&h.Interface, i) }
func (h *SliceHeap[T]) Init()          { heap.Init(&h.Interface) }
func (h *SliceHeap[T]) Pop() T         { return heap.Pop(&h.Interface).(T) }
func (h *SliceHeap[T]) Push(t T)       { heap.Push(&h.Interface, t) }
func (h *SliceHeap[T]) Remove(i int) T { return heap.Remove(&h.Interface, i).(T) }

type Element[T any] struct {
	element *list.Element
}

func (e Element[T]) Unwrap() *list.Element {
	return e.element
}

func (e Element[T]) Ok() bool {
	return e.element != nil
}

func (e Element[T]) Get() T {
	if e.element == nil {
		var zero T
		return zero
	}
	return e.element.Value.(T)
}

func (e Element[T]) Set(t T) {
	e.element.Value = t
}

func (e Element[T]) Next() Element[T] {
	return Element[T]{e.element.Next()}
}

func (e Element[T]) Prev() Element[T] {
	return Element[T]{e.element.Prev()}
}

func (e Element[T]) Forward() iter.Seq[Element[T]] {
	return func(yield func(Element[T]) bool) {
		if !yield(e) {
			return
		}
		for ele := e.Next(); ele.Ok(); ele = ele.Next() {
			if !yield(ele) {
				return
			}
		}
	}
}

func (e Element[T]) ValuesForward() iter.Seq[T] {
	return func(yield func(T) bool) {
		for ele := range e.Forward() {
			if !yield(ele.Get()) {
				return
			}
		}
	}
}

func (e Element[T]) Backward() iter.Seq[Element[T]] {
	return func(yield func(Element[T]) bool) {
		if !yield(e) {
			return
		}
		for ele := e.Prev(); ele.Ok(); ele = ele.Prev() {
			if !yield(ele) {
				return
			}
		}
	}
}

func (e Element[T]) ValuesBackward() iter.Seq[T] {
	return func(yield func(T) bool) {
		for ele := range e.Backward() {
			if !yield(ele.Get()) {
				return
			}
		}
	}
}

type List[T any] struct {
	list *list.List
}

func NewList[T any]() List[T] {
	return List[T]{
		list: list.New(),
	}
}

func (l List[T]) Unwrap() *list.List {
	return l.list
}

func (l List[T]) CloneFunc(cloneT func(T) T) List[T] {
	if l.list == nil {
		return List[T]{}
	}

	new := List[T]{list.New()}
	for ele := l.Front(); ele.element != nil; ele = ele.Next() {
		new.PushBack(cloneT(ele.Get()))
	}
	return new
}

func (l List[T]) Back() Element[T] {
	return Element[T]{element: l.list.Back()}
}

func (l List[T]) Front() Element[T] {
	return Element[T]{l.list.Front()}
}

func (l List[T]) Init() List[T] {
	return List[T]{l.list.Init()}
}

func (l List[T]) InsertAfter(v T, mark Element[T]) Element[T] {
	return Element[T]{l.list.InsertAfter(v, mark.element)}
}

func (l List[T]) InsertBefore(v T, mark Element[T]) Element[T] {
	return Element[T]{l.list.InsertBefore(v, mark.element)}
}

func (l List[T]) Len() int {
	return l.list.Len()
}

func (l List[T]) MoveAfter(e Element[T], mark Element[T]) {
	l.list.MoveAfter(e.element, mark.element)
}

func (l List[T]) MoveBefore(e Element[T], mark Element[T]) {
	l.list.MoveBefore(e.element, mark.element)
}

func (l List[T]) MoveToBack(e Element[T]) {
	l.list.MoveToBack(e.element)
}

func (l List[T]) MoveToFront(e Element[T]) {
	l.list.MoveToFront(e.element)
}

func (l List[T]) PushBack(v T) Element[T] {
	return Element[T]{l.list.PushBack(v)}
}

func (l List[T]) PushBackList(other List[T]) {
	l.list.PushBackList(other.list)
}

func (l List[T]) PushFront(v T) Element[T] {
	return Element[T]{l.list.PushFront(v)}
}

func (l List[T]) PushFrontList(other List[T]) {
	l.list.PushFrontList(other.list)
}

func (l List[T]) Remove(e Element[T]) T {
	return l.list.Remove(e.element).(T)
}

type Ring[T any] struct {
	ring *ring.Ring
}

func NewRing[T any](n int) Ring[T] {
	r := Ring[T]{
		ring: ring.New(n),
	}

	var zero T
	for rr := range r.Forward() {
		rr.ring.Value = zero
	}

	return r
}

func (l Ring[T]) Unwrap() *ring.Ring {
	return l.ring
}

func (l Ring[T]) Get() T {
	return l.ring.Value.(T)
}

func (l Ring[T]) Set(t T) {
	l.ring.Value = t
}

func (r Ring[T]) CloneFunc(cloneT func(T) T) Ring[T] {
	if r.ring == nil {
		return Ring[T]{}
	}
	new := ring.New(r.Len())
	r.Do(func(a T) {
		new.Value = cloneT(a)
		new = new.Next()
	})
	return Ring[T]{new}
}

func (r Ring[T]) Forward() iter.Seq[Ring[T]] {
	return func(yield func(Ring[T]) bool) {
		if !yield(r) {
			return
		}
		for rr := r.Next(); rr.Unwrap() != r.Unwrap(); rr = rr.Next() {
			if !yield(rr) {
				return
			}
		}
	}
}

func (r Ring[T]) Backward() iter.Seq[Ring[T]] {
	return func(yield func(Ring[T]) bool) {
		if !yield(r) {
			return
		}
		for rr := r.Prev(); rr.Unwrap() != r.Unwrap(); rr = rr.Prev() {
			if !yield(rr) {
				return
			}
		}
	}
}

func (r Ring[T]) ValuesForward() iter.Seq[T] {
	return func(yield func(T) bool) {
		if !yield(r.Get()) {
			return
		}
		for rr := r.Next(); rr.Unwrap() != r.Unwrap(); rr = rr.Next() {
			if !yield(rr.Get()) {
				return
			}
		}
	}
}

func (r Ring[T]) ValuesBackward() iter.Seq[T] {
	return func(yield func(T) bool) {
		if !yield(r.Get()) {
			return
		}
		for rr := r.Prev(); rr.Unwrap() != r.Unwrap(); rr = rr.Prev() {
			if !yield(rr.Get()) {
				return
			}
		}
	}
}

func (r Ring[T]) Do(f func(T)) {
	r.ring.Do(func(a any) {
		f(a.(T))
	})
}

func (r Ring[T]) Len() int {
	return r.ring.Len()
}

func (r Ring[T]) Link(s Ring[T]) Ring[T] {
	return Ring[T]{r.ring.Link(s.ring)}
}

func (r Ring[T]) Move(n int) Ring[T] {
	return Ring[T]{r.ring.Move(n)}
}

func (r Ring[T]) Next() Ring[T] {
	return Ring[T]{r.ring.Next()}
}

func (r Ring[T]) Prev() Ring[T] {
	return Ring[T]{r.ring.Prev()}
}

func (r Ring[T]) Unlink(n int) Ring[T] {
	return Ring[T]{r.ring.Unlink(n)}
}
