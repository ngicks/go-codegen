package validatortarget

import (
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

type A []elastic.Elastic[Implementor]

type B map[string]sliceelastic.Elastic[Implementor]

type C [3]option.Option[All]

type D struct {
	Foo  All
	Bar  option.Option[All] `und:"required"`
	FooP *All
	BarP option.Option[*All]   `und:"required"`
	BazP elastic.Elastic[*All] `und:"required,len==3,values:nonnull"`
}
