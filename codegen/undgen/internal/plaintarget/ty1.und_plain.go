package plaintarget

import (
	"bytes"
	"time"

	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub"
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

//undgen:generated
type IncludesImplementorPlain struct {
	Impl         sub.FooPlain[time.Time]
	Opt          sub.Foo[time.Time]               `und:"def"`
	Und          sub.Foo[*bytes.Buffer]           `und:"def"`
	Elastic      []option.Option[sub.Foo[string]] `und:"def"`
	SliceUnd     sub.Foo[int]                     `und:"def"`
	SliceElastic []option.Option[sub.Foo[bool]]   `und:"def"`
}

func (v IncludesImplementor) UndPlain() IncludesImplementorPlain {
	return IncludesImplementorPlain{
		Impl:         v.Impl.UndPlain(),
		Opt:          v.Opt.Value(),
		Und:          v.Und.Value(),
		Elastic:      v.Elastic.Unwrap().Value(),
		SliceUnd:     v.SliceUnd.Value(),
		SliceElastic: v.SliceElastic.Unwrap().Value(),
	}
}

func (v IncludesImplementorPlain) UndRaw() IncludesImplementor {
	return IncludesImplementor{
		Impl:         v.Impl.UndRaw(),
		Opt:          option.Some(v.Opt),
		Und:          und.Defined(v.Und),
		Elastic:      elastic.FromOptions(v.Elastic...),
		SliceUnd:     sliceund.Defined(v.SliceUnd),
		SliceElastic: sliceelastic.FromOptions(v.SliceElastic...),
	}
}
