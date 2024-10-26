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
	SliceElastic sliceelastic.Elastic[sub.Foo[bool]] `und:"def"`
}
