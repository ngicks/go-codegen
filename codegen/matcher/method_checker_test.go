package matcher

import (
	"go/types"
	"testing"

	"gotest.tools/v3/assert"
)

var (
	cyclicConversionMethodsImplementor = `package main

type A struct {}

func (a A) UndPlain() B {
	return B{}
}

type B struct{}

func (b B) UndRaw() A {
	return A{}
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

func (n NotImplementor) UndPlain() B {
	return B{}
}
`
)

func Test(t *testing.T) {
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
