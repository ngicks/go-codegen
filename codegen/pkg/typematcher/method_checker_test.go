package typematcher

import (
	"go/types"
	"testing"

	"github.com/ngicks/go-iterator-helper/hiter"
	"gotest.tools/v3/assert"
)

var cyclicConversionMethodsImplementor = `package main

type A[T any] struct {}

func (a A[T]) UndPlain() B[T] {
	return B[T]{}
}

type B[T any] struct{}

func (b B[T]) UndRaw() A[T] {
	return A[T]{}
}

type AP struct {}

func (a *AP) UndPlain() BP {
	return BP{}
}

type BP struct{}

func (b *BP) UndRaw() AP {
	return AP{}
}

type AI interface {
	UndPlain() BI
}

type BI interface {
	UndRaw() AI
}

type NotImplementor struct{}

func (n NotImplementor) UndPlain() B[any] {
	return B[any]{}
}
`

func TestCyclicConversionMethods(t *testing.T) {
	_, _, pkg := parseStringSource(cyclicConversionMethodsImplementor)

	cmset := CyclicConversionMethods{
		Reverse: "UndRaw",
		Convert: "UndPlain",
	}
	a := pkg.Scope().Lookup("A")
	assert.Assert(t, cmset.IsImplementor(a.Type().(*types.Named)))
	b := pkg.Scope().Lookup("B")
	cmsetRev := cmset
	cmsetRev.From = true
	assert.Assert(t, cmsetRev.IsImplementor(b.Type().(*types.Named)))
	ap := pkg.Scope().Lookup("AP")
	assert.Assert(t, cmset.IsImplementor(ap.Type().(*types.Named)))
	ai := pkg.Scope().Lookup("AI")
	assert.Assert(t, cmset.IsImplementor(ai.Type().(*types.Named)))
	notImplementor := pkg.Scope().Lookup("NotImplementor")
	assert.Assert(t, !cmset.IsImplementor(notImplementor.Type().(*types.Named)))
}

var clonerMethodImplementor = `package main

type C struct{}

func (c C) Clone() C {
	return C{}
}

type CP struct{}

func (c *CP) Clone() CP {
	return CP{}
}

type Param[T, U any] struct{}

func (p Param[T, U]) CloneFunc(cloneT func(T) T, cloneU func(U) U) Param[T, U] {
	return Param[T, U]{}
}

type Total[T, U any] struct {
	C     C
	CP    CP
	CP2  *CP
	p1    Param[T, U]
	p2    Param[string, U]
}
`

func TestClonerMethod(t *testing.T) {
	_, _, pkg := parseStringSource(clonerMethodImplementor)

	method := ClonerMethod{Name: "Clone"}

	c := pkg.Scope().Lookup("C")
	assert.Assert(t, method.IsImplementor(c.Type()))
	cp := pkg.Scope().Lookup("CP")
	assert.Assert(t, method.IsImplementor(cp.Type()))
	param := pkg.Scope().Lookup("Param")
	assert.Assert(t, !method.IsImplementor(param.Type()))
	assert.Assert(t, method.IsFuncImplementor(param.Type()))

	total := pkg.Scope().Lookup("Total").Type().(*types.Named).Underlying().(*types.Struct)
	for i := range hiter.Range(0, 3) {
		assert.Assert(t, method.IsImplementor(total.Field(i).Type()))
	}
	for i := range hiter.Range(3, 4) {
		assert.Assert(t, !method.IsImplementor(total.Field(i).Type()))
		assert.Assert(t, method.IsFuncImplementor(total.Field(i).Type()))
	}
}
