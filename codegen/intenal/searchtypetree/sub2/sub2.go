package sub2

import "github.com/ngicks/und/option"

// implementor
type Foo struct {
}

func (f Foo) UndPlain() FooPlain {
	return FooPlain{}
}

type FooPlain struct {
}

func (f FooPlain) UndRaw() Foo {
	return Foo{}
}

// includes matched type, but external.
type Bar struct {
	O option.Option[string]
}

// not a matched target
type Baz struct {
}
