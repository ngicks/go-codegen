package display

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

type PatchExample struct {
	Foo string
	Bar *int     `json:",omitempty"`
	Baz []string `json:"baz,omitempty"`
}

type Example struct {
	Foo    string
	Bar    option.Option[string]        // no tag
	Baz    option.Option[string]        `und:"def"`
	Qux    und.Und[string]              `und:"def,und"`
	Quux   elastic.Elastic[string]      `und:"null,len==3"`
	Corge  sliceund.Und[string]         `und:"nullish"`
	Grault sliceelastic.Elastic[string] `und:"und,len>=2,values:nonnull"`
}

type Dependent struct {
	Foo Example
	Bar sliceund.Und[Example] `und:"required"`
}
