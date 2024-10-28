package validatortarget

import (
	"fmt"

	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

type All struct {
	Foo    string
	Bar    option.Option[string]        // no tag
	Baz    option.Option[string]        `und:"def"`
	Qux    und.Und[string]              `und:"def,und"`
	Quux   elastic.Elastic[string]      `und:"null,len==3"`
	Corge  sliceund.Und[string]         `und:"nullish"`
	Grault sliceelastic.Elastic[string] `und:"und,len>=2,values:nonnull"`
}

type MapSliceArray struct {
	Foo map[string]option.Option[string] `json:"foo" und:"def"`
	Bar []und.Und[string]                `json:"bar" und:"null"`
	Baz [5]elastic.Elastic[string]       `json:"baz" und:"und,len>=2"`
}

type ContainsImplementor struct {
	I Implementor
	O option.Option[Implementor] `und:"required"`
}

type MapSliceArrayContainsImplementor struct {
	Foo map[string]option.Option[Implementor] `und:"def"`
	Bar []und.Und[Implementor]                `und:"null"`
	Baz [5]elastic.Elastic[Implementor]       `und:"und,len>=2"`
}

//undgen:ignore
type Implementor struct {
	Foo string
}

func (i Implementor) UndValidate() error {
	if i.Foo == "" {
		return fmt.Errorf("huh?")
	}
	return nil
}
