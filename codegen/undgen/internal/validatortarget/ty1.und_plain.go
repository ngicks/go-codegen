package validatortarget

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

//undgen:generated
type AllPlain struct {
	Foo    string
	Bar    option.Option[string]                   // no tag
	Baz    string                                  `und:"def"`
	Qux    option.Option[string]                   `und:"def,und"`
	Quux   option.Option[[3]option.Option[string]] `und:"null,len==3"`
	Corge  option.Option[*struct{}]                `und:"nullish"`
	Grault option.Option[[]string]                 `und:"und,len>=2,values:nonnull"`
}

func (v All) UndPlain() AllPlain {
	return AllPlain{
		Foo: v.Foo,
		Bar: v.Bar,
		Baz: v.Baz.Value(),
		Qux: v.Qux.Unwrap().Value(),
		Quux: und.Map(
			conversion.UnwrapElastic(v.Quux),
			func(o []option.Option[string]) (out [3]option.Option[string]) {
				copy(out[:], o)
				return out
			},
		).Unwrap().Value(),
		Corge:  conversion.UndNullish(v.Corge),
		Grault: conversion.NonNullSlice(conversion.LenNAtLeastSlice(2, conversion.UnwrapElasticSlice(v.Grault))).Unwrap().Value(),
	}
}

func (v AllPlain) UndRaw() All {
	return All{
		Foo: v.Foo,
		Bar: v.Bar,
		Baz: option.Some(v.Baz),
		Qux: conversion.OptionUnd(false, v.Qux),
		Quux: elastic.FromUnd(und.Map(
			conversion.OptionUnd(true, v.Quux),
			func(s [3]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		Corge:  conversion.NullishUndSlice[string](v.Corge),
		Grault: sliceelastic.FromUnd(conversion.NullifySlice(conversion.OptionUndSlice(false, v.Grault))),
	}
}

//undgen:generated
type ContainsImplementorPlain struct {
	I Implementor
	O Implementor `und:"required"`
}

func (v ContainsImplementor) UndPlain() ContainsImplementorPlain {
	return ContainsImplementorPlain{
		I: v.I,
		O: v.O.Value(),
	}
}

func (v ContainsImplementorPlain) UndRaw() ContainsImplementor {
	return ContainsImplementor{
		I: v.I,
		O: option.Some(v.O),
	}
}
