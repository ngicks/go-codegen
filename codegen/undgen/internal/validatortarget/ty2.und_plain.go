package validatortarget

import (
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/option"
)

//undgen:generated
type CPlain [3]option.Option[AllPlain]

func (v C) UndPlain() CPlain {
	return (func(v [3]option.Option[All]) [3]option.Option[AllPlain] {
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

func (v CPlain) UndRaw() C {
	return (func(v [3]option.Option[AllPlain]) [3]option.Option[All] {
		out := [3]option.Option[All]{}

		inner := &out
		for k, v := range v {
			(*inner)[k] = option.Map(
				v,
				conversion.ToRaw,
			)
		}

		return out
	})(v)
}

//undgen:generated
type DPlain struct {
	Foo  AllPlain
	Bar  AllPlain `und:"required"`
	FooP *All
	BarP *All `und:"required"`
}

func (v D) UndPlain() DPlain {
	return DPlain{
		Foo: v.Foo.UndPlain(),
		Bar: option.Map(
			v.Bar,
			conversion.ToPlain,
		).Value(),
		FooP: v.FooP,
		BarP: v.BarP.Value(),
	}
}

func (v DPlain) UndRaw() D {
	return D{
		Foo: v.Foo.UndRaw(),
		Bar: option.Map(
			option.Some(v.Bar),
			conversion.ToRaw,
		),
		FooP: v.FooP,
		BarP: option.Some(v.BarP),
	}
}
