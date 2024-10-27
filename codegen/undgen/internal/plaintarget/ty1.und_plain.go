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
