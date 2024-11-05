package validatortarget

import (
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/option"
)

//undgen:generated
type DPlain struct {
	Foo AllPlain
	Bar AllPlain `und:"required"`
}

func (v D) UndPlain() DPlain {
	return DPlain{
		Foo: v.Foo.UndPlain(),
		Bar: option.Map(
			v.Bar,
			conversion.ToPlain,
		).Value(),
	}
}

func (v DPlain) UndRaw() D {
	return D{
		Foo: v.Foo.UndRaw(),
		Bar: option.Map(
			option.Some(v.Bar),
			conversion.ToRaw,
		),
	}
}
