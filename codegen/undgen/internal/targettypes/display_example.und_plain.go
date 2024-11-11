// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen plain --help
package targettypes

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

//undgen:generated
type ExamplePlain struct {
	Foo   string                `json:"foo"`
	Bar   string                `json:"bar" und:"required"`
	Baz   string                `json:"baz" und:"def"`
	Qux   option.Option[string] `json:"qux" und:"def,null"`
	Quux  [3]option.Option[int] `json:"quux" und:"len==3"`
	Corge []int                 `json:"corge" und:"len>2,values:nonnull"`
}

func (v Example) UndPlain() ExamplePlain {
	return ExamplePlain{
		Foo: v.Foo,
		Bar: v.Bar.Value(),
		Baz: v.Baz.Value(),
		Qux: v.Qux.Unwrap().Value(),
		Quux: sliceund.Map(
			conversion.UnwrapElasticSlice(v.Quux),
			func(o []option.Option[int]) (out [3]option.Option[int]) {
				copy(out[:], o)
				return out
			},
		).Value(),
		Corge: conversion.NonNullSlice(conversion.LenNAtLeastSlice(3, conversion.UnwrapElasticSlice(v.Corge))).Value(),
	}
}

func (v ExamplePlain) UndRaw() Example {
	return Example{
		Foo: v.Foo,
		Bar: option.Some(v.Bar),
		Baz: und.Defined(v.Baz),
		Qux: conversion.OptionUnd(true, v.Qux),
		Quux: sliceelastic.FromUnd(sliceund.Map(
			sliceund.Defined(v.Quux),
			func(s [3]option.Option[int]) []option.Option[int] {
				return s[:]
			},
		)),
		Corge: sliceelastic.FromUnd(conversion.NullifySlice(sliceund.Defined(v.Corge))),
	}
}