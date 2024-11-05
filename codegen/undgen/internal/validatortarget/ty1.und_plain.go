package validatortarget

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
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
