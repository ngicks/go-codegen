package validatortarget

import (
	"github.com/ngicks/und/option"
)

//undgen:generated
type DPlain struct {
	Foo AllPlain
	Bar All `und:"required"`
}

func (v D) UndPlain() DPlain {
	return DPlain{
		Foo: v.Foo.UndPlain(),
		Bar: v.Bar.Value(),
	}
}

func (v DPlain) UndRaw() D {
	return D{
		Foo: v.Foo.UndRaw(),
		Bar: option.Some(v.Bar),
	}
}
