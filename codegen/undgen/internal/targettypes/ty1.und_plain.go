package targettypes

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

//undgen:generated
type AllPlain struct {
	Foo string
	Bar *string
	Baz *struct{}
	Qux []string

	UntouchedOpt      option.Option[int] `json:",omitzero"`
	UntouchedUnd      und.Und[int]       `json:",omitzero"`
	UntouchedSliceUnd sliceund.Und[int]  `json:",omitzero"`

	OptRequired       string                `json:"opt_required,omitzero" und:"required"`
	OptNullish        *struct{}             `json:",omitzero" und:"nullish"`
	OptDef            string                `json:",omitzero" und:"def"`
	OptNull           *struct{}             `json:",omitzero" und:"null"`
	OptUnd            *struct{}             `json:",omitzero" und:"und"`
	OptDefOrUnd       option.Option[string] `json:",omitzero" und:"def,und"`
	OptDefOrNull      option.Option[string] `json:",omitzero" und:"def,null"`
	OptNullOrUnd      *struct{}             `json:",omitzero" und:"null,und"`
	OptDefOrNullOrUnd option.Option[string] `json:",omitzero" und:"def,null,und"`

	UndRequired       string                   `json:",omitzero" und:"required"`
	UndNullish        option.Option[*struct{}] `json:",omitzero" und:"nullish"`
	UndDef            string                   `json:",omitzero" und:"def"`
	UndNull           *struct{}                `json:",omitzero" und:"null"`
	UndUnd            *struct{}                `json:",omitzero" und:"und"`
	UndDefOrUnd       option.Option[string]    `json:",omitzero" und:"def,und"`
	UndDefOrNull      option.Option[string]    `json:",omitzero" und:"def,null"`
	UndNullOrUnd      option.Option[*struct{}] `json:",omitzero" und:"null,und"`
	UndDefOrNullOrUnd und.Und[string]          `json:",omitzero" und:"def,null,und"`

	ElaRequired       []option.Option[string]                `json:",omitzero" und:"required"`
	ElaNullish        option.Option[*struct{}]               `json:",omitzero" und:"nullish"`
	ElaDef            []option.Option[string]                `json:",omitzero" und:"def"`
	ElaNull           *struct{}                              `json:",omitzero" und:"null"`
	ElaUnd            *struct{}                              `json:",omitzero" und:"und"`
	ElaDefOrUnd       option.Option[[]option.Option[string]] `json:",omitzero" und:"def,und"`
	ElaDefOrNull      option.Option[[]option.Option[string]] `json:",omitzero" und:"def,null"`
	ElaNullOrUnd      option.Option[*struct{}]               `json:",omitzero" und:"null,und"`
	ElaDefOrNullOrUnd elastic.Elastic[string]                `json:",omitzero" und:"def,null,und"`

	ElaEqEq option.Option[string]   `json:",omitzero" und:"len==1"`
	ElaGr   []option.Option[string] `json:",omitzero" und:"len>1"`
	ElaGrEq []option.Option[string] `json:",omitzero" und:"len>=1"`
	ElaLe   []option.Option[string] `json:",omitzero" und:"len<1"`
	ElaLeEq []option.Option[string] `json:",omitzero" und:"len<=1"`

	ElaEqEquRequired [2]option.Option[string]                `json:",omitzero" und:"required,len==2"`
	ElaEqEquNullish  und.Und[[2]option.Option[string]]       `json:",omitzero" und:"nullish,len==2"`
	ElaEqEquDef      [2]option.Option[string]                `json:",omitzero" und:"def,len==2"`
	ElaEqEquNull     option.Option[[2]option.Option[string]] `json:",omitzero" und:"null,len==2"`
	ElaEqEquUnd      option.Option[[2]option.Option[string]] `json:",omitzero" und:"und,len==2"`

	ElaEqEqNonNullSlice      und.Und[[]string]        `json:",omitzero" und:"values:nonnull"`
	ElaEqEqNonNullNullSlice  *struct{}                `json:",omitzero" und:"null,values:nonnull"`
	ElaEqEqNonNullSingle     string                   `json:",omitzero" und:"values:nonnull,len==1"`
	ElaEqEqNonNullNullSingle option.Option[string]    `json:",omitzero" und:"null,values:nonnull,len==1"`
	ElaEqEqNonNull           [3]string                `json:",omitzero" und:"values:nonnull,len==3"`
	ElaEqEqNonNullNull       option.Option[[3]string] `json:",omitzero" und:"null,values:nonnull,len==3"`
}

func (v All) UndPlain() AllPlain {
	return AllPlain{
		Foo:               v.Foo,
		Bar:               v.Bar,
		Baz:               v.Baz,
		Qux:               v.Qux,
		UntouchedOpt:      v.UntouchedOpt,
		UntouchedUnd:      v.UntouchedUnd,
		UntouchedSliceUnd: v.UntouchedSliceUnd,
		OptRequired:       v.OptRequired.Value(),
		OptNullish:        nil,
		OptDef:            v.OptDef.Value(),
		OptNull:           nil,
		OptUnd:            nil,
		OptDefOrUnd:       v.OptDefOrUnd,
		OptDefOrNull:      v.OptDefOrNull,
		OptNullOrUnd:      nil,
		OptDefOrNullOrUnd: v.OptDefOrNullOrUnd,
		UndRequired:       v.UndRequired.Value(),
		UndNullish:        conversion.UndNullish(v.UndNullish),
		UndDef:            v.UndDef.Value(),
		UndNull:           nil,
		UndUnd:            nil,
		UndDefOrUnd:       v.UndDefOrUnd.Unwrap().Value(),
		UndDefOrNull:      v.UndDefOrNull.Unwrap().Value(),
		UndNullOrUnd:      conversion.UndNullish(v.UndNullOrUnd),
		UndDefOrNullOrUnd: v.UndDefOrNullOrUnd,
		ElaRequired:       v.ElaRequired.Unwrap().Value(),
		ElaNullish:        conversion.UndNullish(v.ElaNullish),
		ElaDef:            v.ElaDef.Unwrap().Value(),
		ElaNull:           nil,
		ElaUnd:            nil,
		ElaDefOrUnd:       conversion.UnwrapElastic(v.ElaDefOrUnd).Unwrap().Value(),
		ElaDefOrNull:      conversion.UnwrapElastic(v.ElaDefOrNull).Unwrap().Value(),
		ElaNullOrUnd:      conversion.UndNullish(v.ElaNullOrUnd),
		ElaDefOrNullOrUnd: v.ElaDefOrNullOrUnd,
		ElaEqEq: conversion.UnwrapLen1(und.Map(
			conversion.UnwrapElastic(v.ElaEqEq),
			func(o []option.Option[string]) (out [1]option.Option[string]) {
				copy(out[:], o)
				return out
			},
		)).Value(),
		ElaGr:   conversion.LenNAtLeast(2, conversion.UnwrapElastic(v.ElaGr)).Value(),
		ElaGrEq: conversion.LenNAtLeast(1, conversion.UnwrapElastic(v.ElaGrEq)).Value(),
		ElaLe:   conversion.LenNAtMost(0, conversion.UnwrapElastic(v.ElaLe)).Value(),
		ElaLeEq: conversion.LenNAtMost(1, conversion.UnwrapElastic(v.ElaLeEq)).Value(),
		ElaEqEquRequired: und.Map(
			conversion.UnwrapElastic(v.ElaEqEquRequired),
			func(o []option.Option[string]) (out [2]option.Option[string]) {
				copy(out[:], o)
				return out
			},
		).Value(),
		ElaEqEquNullish: und.Map(
			conversion.UnwrapElastic(v.ElaEqEquNullish),
			func(o []option.Option[string]) (out [2]option.Option[string]) {
				copy(out[:], o)
				return out
			},
		),
		ElaEqEquDef: und.Map(
			conversion.UnwrapElastic(v.ElaEqEquDef),
			func(o []option.Option[string]) (out [2]option.Option[string]) {
				copy(out[:], o)
				return out
			},
		).Value(),
		ElaEqEquNull: und.Map(
			conversion.UnwrapElastic(v.ElaEqEquNull),
			func(o []option.Option[string]) (out [2]option.Option[string]) {
				copy(out[:], o)
				return out
			},
		).Unwrap().Value(),
		ElaEqEquUnd: und.Map(
			conversion.UnwrapElastic(v.ElaEqEquUnd),
			func(o []option.Option[string]) (out [2]option.Option[string]) {
				copy(out[:], o)
				return out
			},
		).Unwrap().Value(),
		ElaEqEqNonNullSlice:     conversion.NonNull(conversion.UnwrapElastic(v.ElaEqEqNonNullSlice)),
		ElaEqEqNonNullNullSlice: nil,
		ElaEqEqNonNullSingle: conversion.UnwrapLen1(und.Map(
			und.Map(
				conversion.UnwrapElastic(v.ElaEqEqNonNullSingle),
				func(o []option.Option[string]) (out [1]option.Option[string]) {
					copy(out[:], o)
					return out
				},
			),
			func(s [1]option.Option[string]) (r [1]string) {
				for i := 0; i < 1; i++ {
					r[i] = s[i].Value()
				}
				return
			},
		)).Value(),
		ElaEqEqNonNullNullSingle: conversion.UnwrapLen1(und.Map(
			und.Map(
				conversion.UnwrapElastic(v.ElaEqEqNonNullNullSingle),
				func(o []option.Option[string]) (out [1]option.Option[string]) {
					copy(out[:], o)
					return out
				},
			),
			func(s [1]option.Option[string]) (r [1]string) {
				for i := 0; i < 1; i++ {
					r[i] = s[i].Value()
				}
				return
			},
		)).Unwrap().Value(),
		ElaEqEqNonNull: und.Map(
			und.Map(
				conversion.UnwrapElastic(v.ElaEqEqNonNull),
				func(o []option.Option[string]) (out [3]option.Option[string]) {
					copy(out[:], o)
					return out
				},
			),
			func(s [3]option.Option[string]) (r [3]string) {
				for i := 0; i < 3; i++ {
					r[i] = s[i].Value()
				}
				return
			},
		).Value(),
		ElaEqEqNonNullNull: und.Map(
			und.Map(
				conversion.UnwrapElastic(v.ElaEqEqNonNullNull),
				func(o []option.Option[string]) (out [3]option.Option[string]) {
					copy(out[:], o)
					return out
				},
			),
			func(s [3]option.Option[string]) (r [3]string) {
				for i := 0; i < 3; i++ {
					r[i] = s[i].Value()
				}
				return
			},
		).Unwrap().Value(),
	}
}

func (v AllPlain) UndRaw() All {
	return All{
		Foo:               v.Foo,
		Bar:               v.Bar,
		Baz:               v.Baz,
		Qux:               v.Qux,
		UntouchedOpt:      v.UntouchedOpt,
		UntouchedUnd:      v.UntouchedUnd,
		UntouchedSliceUnd: v.UntouchedSliceUnd,
		OptRequired:       option.Some(v.OptRequired),
		OptNullish:        option.None[string](),
		OptDef:            option.Some(v.OptDef),
		OptNull:           option.None[string](),
		OptUnd:            option.None[string](),
		OptDefOrUnd:       v.OptDefOrUnd,
		OptDefOrNull:      v.OptDefOrNull,
		OptNullOrUnd:      option.None[string](),
		OptDefOrNullOrUnd: v.OptDefOrNullOrUnd,
		UndRequired:       und.Defined(v.UndRequired),
		UndNullish:        conversion.NullishUnd[string](v.UndNullish),
		UndDef:            und.Defined(v.UndDef),
		UndNull:           und.Null[string](),
		UndUnd:            und.Undefined[string](),
		UndDefOrUnd:       conversion.OptionUnd(false, v.UndDefOrUnd),
		UndDefOrNull:      conversion.OptionUnd(true, v.UndDefOrNull),
		UndNullOrUnd:      conversion.NullishUnd[string](v.UndNullOrUnd),
		UndDefOrNullOrUnd: v.UndDefOrNullOrUnd,
		ElaRequired:       elastic.FromOptions(v.ElaRequired...),
		ElaNullish:        conversion.NullishElastic[string](v.ElaNullish),
		ElaDef:            elastic.FromOptions(v.ElaDef...),
		ElaNull:           elastic.Null[string](),
		ElaUnd:            elastic.Undefined[string](),
		ElaDefOrUnd:       conversion.OptionOptionElastic(false, v.ElaDefOrUnd),
		ElaDefOrNull:      conversion.OptionOptionElastic(true, v.ElaDefOrNull),
		ElaNullOrUnd:      conversion.NullishElastic[string](v.ElaNullOrUnd),
		ElaDefOrNullOrUnd: v.ElaDefOrNullOrUnd,
		ElaEqEq: elastic.FromUnd(und.Map(
			conversion.WrapLen1(und.Defined(v.ElaEqEq)),
			func(s [1]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaGr:   elastic.FromUnd(und.Defined(v.ElaGr)),
		ElaGrEq: elastic.FromUnd(und.Defined(v.ElaGrEq)),
		ElaLe:   elastic.FromUnd(und.Defined(v.ElaLe)),
		ElaLeEq: elastic.FromUnd(und.Defined(v.ElaLeEq)),
		ElaEqEquRequired: elastic.FromUnd(und.Map(
			und.Defined(v.ElaEqEquRequired),
			func(s [2]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaEqEquNullish: elastic.FromUnd(und.Map(
			v.ElaEqEquNullish,
			func(s [2]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaEqEquDef: elastic.FromUnd(und.Map(
			und.Defined(v.ElaEqEquDef),
			func(s [2]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaEqEquNull: elastic.FromUnd(und.Map(
			conversion.OptionUnd(true, v.ElaEqEquNull),
			func(s [2]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaEqEquUnd: elastic.FromUnd(und.Map(
			conversion.OptionUnd(false, v.ElaEqEquUnd),
			func(s [2]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaEqEqNonNullSlice:     elastic.FromUnd(conversion.Nullify(v.ElaEqEqNonNullSlice)),
		ElaEqEqNonNullNullSlice: elastic.Null[string](),
		ElaEqEqNonNullSingle: elastic.FromUnd(und.Map(
			und.Map(
				conversion.WrapLen1(und.Defined(v.ElaEqEqNonNullSingle)),
				func(s [1]string) (out [1]option.Option[string]) {
					for i := 0; i < 1; i++ {
						out[i] = option.Some(s[i])
					}
					return
				},
			),
			func(s [1]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaEqEqNonNullNullSingle: elastic.FromUnd(und.Map(
			und.Map(
				conversion.WrapLen1(conversion.OptionUnd(true, v.ElaEqEqNonNullNullSingle)),
				func(s [1]string) (out [1]option.Option[string]) {
					for i := 0; i < 1; i++ {
						out[i] = option.Some(s[i])
					}
					return
				},
			),
			func(s [1]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaEqEqNonNull: elastic.FromUnd(und.Map(
			und.Map(
				und.Defined(v.ElaEqEqNonNull),
				func(s [3]string) (out [3]option.Option[string]) {
					for i := 0; i < 3; i++ {
						out[i] = option.Some(s[i])
					}
					return
				},
			),
			func(s [3]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
		ElaEqEqNonNullNull: elastic.FromUnd(und.Map(
			und.Map(
				conversion.OptionUnd(true, v.ElaEqEqNonNullNull),
				func(s [3]string) (out [3]option.Option[string]) {
					for i := 0; i < 3; i++ {
						out[i] = option.Some(s[i])
					}
					return
				},
			),
			func(s [3]option.Option[string]) []option.Option[string] {
				return s[:]
			},
		)),
	}
}

//undgen:generated
type WithTypeParamPlain[T any] struct {
	Foo string
	Bar T
	Baz T `json:",omitzero" und:"required"`
}

func (v WithTypeParam[T]) UndPlain() WithTypeParamPlain[T] {
	return WithTypeParamPlain[T]{
		Foo: v.Foo,
		Bar: v.Bar,
		Baz: v.Baz.Value(),
	}
}

func (v WithTypeParamPlain[T]) UndRaw() WithTypeParam[T] {
	return WithTypeParam[T]{
		Foo: v.Foo,
		Bar: v.Bar,
		Baz: option.Some(v.Baz),
	}
}
