package builtins

import "unsafe"

type Aaaa unsafe.Pointer

type BuiltIns[T any] struct {
	T   T
	Foo string
	Bar map[string]int
	Err error
}
