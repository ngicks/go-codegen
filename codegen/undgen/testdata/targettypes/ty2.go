package targettypes

import (
	"github.com/ngicks/go-codegen/codegen/undgen/testdata/targettypes/sub"
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceElastic "github.com/ngicks/und/sliceund/elastic"
)

type A struct {
	A option.Option[string]
}

type B struct {
	B und.Und[int]
}

type C []elastic.Elastic[string]

type D [5]sliceund.Und[string]

type F map[string]sliceElastic.Elastic[string]

type Parametrized[T any] struct {
	A option.Option[T]
}

type NotATarget struct {
	Foo string
	Bar int
	Baz map[option.Option[string]]bool
}

type IncludesSubTarget struct {
	Foo sub.Baz[string]
}

type IncludesImplementor struct {
	Foo sub.Foo[string]
}

type NestedImplementor struct {
	Foo option.Option[sub.Foo[string]]
}

type NestedImplementor2 struct {
	Foo sub.IncludesImplementor
}
