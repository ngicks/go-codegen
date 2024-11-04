package validatortarget

import (
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/option"
)

//undgen:generated
type AllPlain struct {
	Foo    string
	Bar    option.Option[string]                   // no tag
	Baz    string                                  `und:"def"`
	Qux    option.Option[string]                   `und:"def,und"`
	Quux   option.Option[[3]option.Option[string]] `und:"null,len==3"`
	Corge  option.Option[conversion.Empty]         `und:"nullish"`
	Grault option.Option[[]string]                 `und:"und,len>=2,values:nonnull"`
}

//undgen:generated
type MapSliceArrayPlain struct {
	Foo map[string]string                         `json:"foo" und:"def"`
	Bar []conversion.Empty                        `json:"bar" und:"null"`
	Baz [5]option.Option[[]option.Option[string]] `json:"baz" und:"und,len>=2"`
}

//undgen:generated
type ContainsImplementorPlain struct {
	I Implementor
	O Implementor `und:"required"`
}

//undgen:generated
type MapSliceArrayContainsImplementorPlain struct {
	Foo map[string]Implementor                         `und:"def"`
	Bar []conversion.Empty                             `und:"null"`
	Baz [5]option.Option[[]option.Option[Implementor]] `und:"und,len>=2"`
}
