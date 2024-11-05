package plaintarget

import (
	"bytes"
	"time"

	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub"
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

//undgen:generated
type IncludesImplementorPlain struct {
	Impl         sub.FooPlain[time.Time]
	Opt          sub.FooPlain[time.Time]               `und:"def"`
	Und          sub.FooPlain[*bytes.Buffer]           `und:"def"`
	Elastic      []option.Option[sub.FooPlain[string]] `und:"def"`
	SliceUnd     sub.FooPlain[int]                     `und:"def"`
	SliceElastic [2]option.Option[sub.FooPlain[bool]]  `und:"len==2"`
}

func (v IncludesImplementor) UndPlain() IncludesImplementorPlain {
	return IncludesImplementorPlain{
		Impl: v.Impl.UndPlain(),
		Opt: option.Map(
			v.Opt,
			conversion.ToPlain,
		).Value(),
		Und: und.Map(
			v.Und,
			conversion.ToPlain,
		).Value(),
		Elastic: elastic.Map(
			v.Elastic,
			conversion.ToPlain,
		).Unwrap().Value(),
		SliceUnd: sliceund.Map(
			v.SliceUnd,
			conversion.ToPlain,
		).Value(),
		SliceElastic: sliceund.Map(
			conversion.UnwrapElasticSlice(sliceelastic.Map(
				v.SliceElastic,
				conversion.ToPlain,
			)),
			func(o []option.Option[sub.FooPlain[bool]]) (out [2]option.Option[sub.FooPlain[bool]]) {
				copy(out[:], o)
				return out
			},
		).Value(),
	}
}

func (v IncludesImplementorPlain) UndRaw() IncludesImplementor {
	return IncludesImplementor{
		Impl: v.Impl.UndRaw(),
		Opt: option.Map(
			option.Some(v.Opt),
			conversion.ToRaw,
		),
		Und: und.Map(
			und.Defined(v.Und),
			conversion.ToRaw,
		),
		Elastic: elastic.Map(
			elastic.FromOptions(v.Elastic...),
			conversion.ToRaw,
		),
		SliceUnd: sliceund.Map(
			sliceund.Defined(v.SliceUnd),
			conversion.ToRaw,
		),
		SliceElastic: sliceelastic.Map(
			sliceelastic.FromUnd(sliceund.Map(
				sliceund.Defined(v.SliceElastic),
				func(s [2]option.Option[sub.FooPlain[bool]]) []option.Option[sub.FooPlain[bool]] {
					return s[:]
				},
			)),
			conversion.ToRaw,
		),
	}
}

//undgen:generated
type IncludesImplementorArraySliceMapPlain struct {
	A1 [3]sub.FooPlain[time.Time]
	A2 [5]sub.FooPlain[time.Time]                      `und:"def"`
	S1 []sub.FooPlain[*bytes.Buffer]                   `und:"def"`
	S2 [][]option.Option[sub.FooPlain[string]]         `und:"def,len<3"`
	M1 map[string]sub.FooPlain[int]                    `und:"def"`
	M2 map[string][2]option.Option[sub.FooPlain[bool]] `und:"len==2"`
}

func (v IncludesImplementorArraySliceMap) UndPlain() IncludesImplementorArraySliceMapPlain {
	return IncludesImplementorArraySliceMapPlain{
		A1: func(in [3]sub.Foo[time.Time]) (out [3]sub.FooPlain[time.Time]) {
			for k, v := range in {
				out[k] = v.UndPlain()
			}
			return out
		}(v.A1),
		A2: func(in [5]option.Option[sub.Foo[time.Time]]) (out [5]sub.FooPlain[time.Time]) {
			for k, v := range in {
				out[k] = option.Map(
					v,
					conversion.ToPlain,
				).Value()
			}
			return out
		}(v.A2),
		S1: func(in []und.Und[sub.Foo[*bytes.Buffer]]) []sub.FooPlain[*bytes.Buffer] {
			out := make([]sub.FooPlain[*bytes.Buffer], len(in))
			for k, v := range in {
				out[k] = und.Map(
					v,
					conversion.ToPlain,
				).Value()
			}
			return out
		}(v.S1),
		S2: func(in []elastic.Elastic[sub.Foo[string]]) [][]option.Option[sub.FooPlain[string]] {
			out := make([][]option.Option[sub.FooPlain[string]], len(in))
			for k, v := range in {
				out[k] = conversion.LenNAtMost(2, conversion.UnwrapElastic(elastic.Map(
					v,
					conversion.ToPlain,
				))).Value()
			}
			return out
		}(v.S2),
		M1: func(in map[string]sliceund.Und[sub.Foo[int]]) map[string]sub.FooPlain[int] {
			out := make(map[string]sub.FooPlain[int], len(in))
			for k, v := range in {
				out[k] = sliceund.Map(
					v,
					conversion.ToPlain,
				).Value()
			}
			return out
		}(v.M1),
		M2: func(in map[string]sliceelastic.Elastic[sub.Foo[bool]]) map[string][2]option.Option[sub.FooPlain[bool]] {
			out := make(map[string][2]option.Option[sub.FooPlain[bool]], len(in))
			for k, v := range in {
				out[k] = sliceund.Map(
					conversion.UnwrapElasticSlice(sliceelastic.Map(
						v,
						conversion.ToPlain,
					)),
					func(o []option.Option[sub.FooPlain[bool]]) (out [2]option.Option[sub.FooPlain[bool]]) {
						copy(out[:], o)
						return out
					},
				).Value()
			}
			return out
		}(v.M2),
	}
}

func (v IncludesImplementorArraySliceMapPlain) UndRaw() IncludesImplementorArraySliceMap {
	return IncludesImplementorArraySliceMap{
		A1: func(in [3]sub.FooPlain[time.Time]) (out [3]sub.Foo[time.Time]) {
			for k, v := range in {
				out[k] = v.UndRaw()
			}
			return out
		}(v.A1),
		A2: func(in [5]sub.FooPlain[time.Time]) (out [5]option.Option[sub.Foo[time.Time]]) {
			for k, v := range in {
				out[k] = option.Map(
					option.Some(v),
					conversion.ToRaw,
				)
			}
			return out
		}(v.A2),
		S1: func(in []sub.FooPlain[*bytes.Buffer]) []und.Und[sub.Foo[*bytes.Buffer]] {
			out := make([]und.Und[sub.Foo[*bytes.Buffer]], len(in))
			for k, v := range in {
				out[k] = und.Map(
					und.Defined(v),
					conversion.ToRaw,
				)
			}
			return out
		}(v.S1),
		S2: func(in [][]option.Option[sub.FooPlain[string]]) []elastic.Elastic[sub.Foo[string]] {
			out := make([]elastic.Elastic[sub.Foo[string]], len(in))
			for k, v := range in {
				out[k] = elastic.Map(
					elastic.FromUnd(und.Defined(v)),
					conversion.ToRaw,
				)
			}
			return out
		}(v.S2),
		M1: func(in map[string]sub.FooPlain[int]) map[string]sliceund.Und[sub.Foo[int]] {
			out := make(map[string]sliceund.Und[sub.Foo[int]], len(in))
			for k, v := range in {
				out[k] = sliceund.Map(
					sliceund.Defined(v),
					conversion.ToRaw,
				)
			}
			return out
		}(v.M1),
		M2: func(in map[string][2]option.Option[sub.FooPlain[bool]]) map[string]sliceelastic.Elastic[sub.Foo[bool]] {
			out := make(map[string]sliceelastic.Elastic[sub.Foo[bool]], len(in))
			for k, v := range in {
				out[k] = sliceelastic.Map(
					sliceelastic.FromUnd(sliceund.Map(
						sliceund.Defined(v),
						func(s [2]option.Option[sub.FooPlain[bool]]) []option.Option[sub.FooPlain[bool]] {
							return s[:]
						},
					)),
					conversion.ToRaw,
				)
			}
			return out
		}(v.M2),
	}
}

//undgen:generated
type WrappedPlain map[string][3][]sub.FooPlain[string]

func (v Wrapped) UndPlain() WrappedPlain {
	return (func(v Wrapped) map[string][3][]sub.FooPlain[string] {
		out := make(map[string][3][]sub.FooPlain[string], len(v))

		inner := out
		for k, v := range v {
			outer := &inner
			inner := [3][]sub.FooPlain[string]{}
			for k, v := range v {
				outer := &inner
				inner := make([]sub.FooPlain[string], len(v))
				for k, v := range v {
					inner[k] = v.UndPlain()
				}
				(*outer)[k] = inner
			}
			(*outer)[k] = inner
		}

		return out
	})(v)
}
