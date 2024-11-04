package plaintarget

import (
	"bytes"
	"time"

	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub"
	"github.com/ngicks/und/option"
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

//undgen:generated
type IncludesImplementorArraySliceMapPlain struct {
	A1 [3]sub.FooPlain[time.Time]
	A2 [5]sub.FooPlain[time.Time]                      `und:"def"`
	S1 []sub.FooPlain[*bytes.Buffer]                   `und:"def"`
	S2 [][]option.Option[sub.FooPlain[string]]         `und:"def,len<3"`
	M1 map[string]sub.FooPlain[int]                    `und:"def"`
	M2 map[string][2]option.Option[sub.FooPlain[bool]] `und:"len==2"`
}
