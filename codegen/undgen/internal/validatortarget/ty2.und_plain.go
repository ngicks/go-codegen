package validatortarget

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/elastic"
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
				func(v All) AllPlain {
					vv := v.UndPlain()
					return vv
				},
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
				func(v AllPlain) All {
					vv := v.UndRaw()
					return vv
				},
			)
		}

		return out
	})(v)
}

//undgen:generated
type DPlain struct {
	Foo  AllPlain
	Bar  AllPlain `und:"required"`
	FooP *AllPlain
	BarP *AllPlain    `und:"required"`
	BazP [3]*AllPlain `und:"required,len==3,values:nonnull"`
}

func (v D) UndPlain() DPlain {
	return DPlain{
		Foo: v.Foo.UndPlain(),
		Bar: option.Map(
			v.Bar,
			func(v All) AllPlain {
				vv := v.UndPlain()
				return vv
			},
		).Value(),
		FooP: func(v *All) *AllPlain {
			if v == nil {
				return nil
			}
			vv := v.UndPlain()
			return &vv
		}(v.FooP),
		BarP: option.Map(
			v.BarP,
			func(v *All) *AllPlain {
				if v == nil {
					return nil
				}
				vv := v.UndPlain()
				return &vv
			},
		).Value(),
		BazP: und.Map(
			und.Map(
				conversion.UnwrapElastic(elastic.Map(
					v.BazP,
					func(v *All) *AllPlain {
						if v == nil {
							return nil
						}
						vv := v.UndPlain()
						return &vv
					},
				)),
				func(o []option.Option[*AllPlain]) (out [3]option.Option[*AllPlain]) {
					copy(out[:], o)
					return out
				},
			),
			func(s [3]option.Option[*AllPlain]) (r [3]*AllPlain) {
				for i := 0; i < 3; i++ {
					r[i] = s[i].Value()
				}
				return
			},
		).Value(),
	}
}

func (v DPlain) UndRaw() D {
	return D{
		Foo: v.Foo.UndRaw(),
		Bar: option.Map(
			option.Some(v.Bar),
			func(v AllPlain) All {
				vv := v.UndRaw()
				return vv
			},
		),
		FooP: func(v *AllPlain) *All {
			if v == nil {
				return nil
			}
			vv := v.UndRaw()
			return &vv
		}(v.FooP),
		BarP: option.Map(
			option.Some(v.BarP),
			func(v *AllPlain) *All {
				if v == nil {
					return nil
				}
				vv := v.UndRaw()
				return &vv
			},
		),
		BazP: elastic.Map(
			elastic.FromUnd(und.Map(
				und.Map(
					und.Defined(v.BazP),
					func(s [3]*AllPlain) (out [3]option.Option[*AllPlain]) {
						for i := 0; i < 3; i++ {
							out[i] = option.Some(s[i])
						}
						return
					},
				),
				func(s [3]option.Option[*AllPlain]) []option.Option[*AllPlain] {
					return s[:]
				},
			)),
			func(v *AllPlain) *All {
				if v == nil {
					return nil
				}
				vv := v.UndRaw()
				return &vv
			},
		),
	}
}
