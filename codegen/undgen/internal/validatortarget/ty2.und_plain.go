package validatortarget

import (
	"github.com/ngicks/und/option"
)

//undgen:generated
type DPlain struct {
	Foo All
	Bar All `und:"required"`
}

func (v D) UndPlain() DPlain {
	return DPlain{
		Foo: v.Foo,
		Bar: v.Bar.Value(),
	}
}

func (v DPlain) UndRaw() D {
	return D{
		Foo: v.Foo,
		Bar: option.Some(v.Bar),
	}
}
