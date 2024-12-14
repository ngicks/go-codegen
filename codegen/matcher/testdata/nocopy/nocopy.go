package nocopy

import (
	"sync"
	"sync/atomic"
)

func init() {
	a := A{}
	a2 := a // ok, pointer

	b := B{}
	b2 := b // ok, interface is a pointer

	c := C{}
	c2 := c // ok, pointer. But is a no copy object.

	d := D{}
	d2 := d
	// assignment copies lock value to d2: github.com/ngicks/go-codegen/codegen/matcher.D
	// contains sync/atomic.Int64 contains sync/atomic.noCopycopylocksdefault

	e := E{}
	e2 := e // same, vet warning

	g := G{}
	g2 := g // same, vet warning

	h := H{}
	h2 := h // ok

	discard(a2, b2, c2, d2, e2, g2, h2)
}

func discard(args ...any) {}

type A struct {
	mu *sync.Mutex
}

type B struct {
	locker sync.Locker
}

type C struct {
	sync.Locker
}

type D struct {
	a atomic.Int64
}

type E [5]atomic.Int64

type F [3][5]atomic.Int64

type G struct {
	E E
}

type H map[string]atomic.Int64

type Tree struct {
	L, R *Tree
}
