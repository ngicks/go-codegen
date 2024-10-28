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

type IncludesImplementor struct {
	Impl         sub.Foo[time.Time]
	Opt          option.Option[sub.Foo[time.Time]]   `und:"def"`
	Und          und.Und[sub.Foo[*bytes.Buffer]]     `und:"def"`
	Elastic      elastic.Elastic[sub.Foo[string]]    `und:"def"`
	SliceUnd     sliceund.Und[sub.Foo[int]]          `und:"def"`
	SliceElastic sliceelastic.Elastic[sub.Foo[bool]] `und:"len==2"`
}

type IncludesImplementorArraySliceMap struct {
	A1 [3]sub.Foo[time.Time]
	A2 [5]option.Option[sub.Foo[time.Time]]           `und:"def"`
	S1 []und.Und[sub.Foo[*bytes.Buffer]]              `und:"def"`
	S2 []elastic.Elastic[sub.Foo[string]]             `und:"def,len<3"`
	M1 map[string]sliceund.Und[sub.Foo[int]]          `und:"def"`
	M2 map[string]sliceelastic.Elastic[sub.Foo[bool]] `und:"len==2"`
}
