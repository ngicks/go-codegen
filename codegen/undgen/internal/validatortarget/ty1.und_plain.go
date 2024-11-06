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
	Corge  option.Option[conversion.Empty]         `und:"nullish"`
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
type MapSliceArrayPlain struct {
	Foo map[string]string                         `json:"foo" und:"def"`
	Bar []conversion.Empty                        `json:"bar" und:"null"`
	Baz [5]option.Option[[]option.Option[string]] `json:"baz" und:"und,len>=2"`
}

func (v MapSliceArray) UndPlain() MapSliceArrayPlain {
	return MapSliceArrayPlain{
		Foo: (func(v map[string]option.Option[string]) map[string]string {
			out := make(map[string]string, len(v))

			inner := &out
			for k, v := range v {
				(*inner)[k] = v.Value()
			}

			return out
		})(v.Foo),
		Bar: (func(v []und.Und[string]) []conversion.Empty {
			out := make([]conversion.Empty, len(v))

			inner := &out
			for k, v := range v {
				(*inner)[k] = nil
				_ = v // just to avoid compilation error
			}

			return out
		})(v.Bar),
		Baz: (func(v [5]elastic.Elastic[string]) [5]option.Option[[]option.Option[string]] {
			out := [5]option.Option[[]option.Option[string]]{}

			inner := &out
			for k, v := range v {
				(*inner)[k] = conversion.LenNAtLeast(2, conversion.UnwrapElastic(v)).Unwrap().Value()
			}

			return out
		})(v.Baz),
	}
}

func (v MapSliceArrayPlain) UndRaw() MapSliceArray {
	return MapSliceArray{
		Foo: (func(v map[string]string) map[string]option.Option[string] {
			out := make(map[string]option.Option[string], len(v))

			inner := &out
			for k, v := range v {
				(*inner)[k] = option.Some(v)
			}

			return out
		})(v.Foo),
		Bar: (func(v []conversion.Empty) []und.Und[string] {
			out := make([]und.Und[string], len(v))

			inner := &out
			for k, v := range v {
				(*inner)[k] = und.Null[string]()
				_ = v // just to avoid compilation error
			}

			return out
		})(v.Bar),
		Baz: (func(v [5]option.Option[[]option.Option[string]]) [5]elastic.Elastic[string] {
			out := [5]elastic.Elastic[string]{}

			inner := &out
			for k, v := range v {
				(*inner)[k] = elastic.FromUnd(conversion.OptionUnd(false, v))
			}

			return out
		})(v.Baz),
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

//undgen:generated
type MapSliceArrayContainsImplementorPlain struct {
	Foo map[string]Implementor                         `und:"def"`
	Bar []conversion.Empty                             `und:"null"`
	Baz [5]option.Option[[]option.Option[Implementor]] `und:"und,len>=2"`
}

func (v MapSliceArrayContainsImplementor) UndPlain() MapSliceArrayContainsImplementorPlain {
	return MapSliceArrayContainsImplementorPlain{
		Foo: (func(v map[string]option.Option[Implementor]) map[string]Implementor {
			out := make(map[string]Implementor, len(v))

			inner := &out
			for k, v := range v {
				(*inner)[k] = v.Value()
			}

			return out
		})(v.Foo),
		Bar: (func(v []und.Und[Implementor]) []conversion.Empty {
			out := make([]conversion.Empty, len(v))

			inner := &out
			for k, v := range v {
				(*inner)[k] = nil
				_ = v // just to avoid compilation error
			}

			return out
		})(v.Bar),
		Baz: (func(v [5]elastic.Elastic[Implementor]) [5]option.Option[[]option.Option[Implementor]] {
			out := [5]option.Option[[]option.Option[Implementor]]{}

			inner := &out
			for k, v := range v {
				(*inner)[k] = conversion.LenNAtLeast(2, conversion.UnwrapElastic(v)).Unwrap().Value()
			}

			return out
		})(v.Baz),
	}
}

func (v MapSliceArrayContainsImplementorPlain) UndRaw() MapSliceArrayContainsImplementor {
	return MapSliceArrayContainsImplementor{
		Foo: (func(v map[string]Implementor) map[string]option.Option[Implementor] {
			out := make(map[string]option.Option[Implementor], len(v))

			inner := &out
			for k, v := range v {
				(*inner)[k] = option.Some(v)
			}

			return out
		})(v.Foo),
		Bar: (func(v []conversion.Empty) []und.Und[Implementor] {
			out := make([]und.Und[Implementor], len(v))

			inner := &out
			for k, v := range v {
				(*inner)[k] = und.Null[Implementor]()
				_ = v // just to avoid compilation error
			}

			return out
		})(v.Bar),
		Baz: (func(v [5]option.Option[[]option.Option[Implementor]]) [5]elastic.Elastic[Implementor] {
			out := [5]elastic.Elastic[Implementor]{}

			inner := &out
			for k, v := range v {
				(*inner)[k] = elastic.FromUnd(conversion.OptionUnd(false, v))
			}

			return out
		})(v.Baz),
	}
}
