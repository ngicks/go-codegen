package validatortarget

import (
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/option"
)

//undgen:generated
type CPlain [3]option.Option[AllPlain]

func (v C) UndPlain() CPlain {
	return (func(v C) [3]option.Option[AllPlain] {
		out := [3]option.Option[AllPlain]{}

		inner := &out
		for k, v := range v {
			(*inner)[k] = option.Map(
				v,
				conversion.ToPlain,
			)
		}

		return out
	})(v)
}

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
