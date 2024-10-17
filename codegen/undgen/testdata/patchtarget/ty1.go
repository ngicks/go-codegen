package patchtarget

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
)

type All struct {
	Foo string
	Bar *int
	Baz *struct{}
	Qux []string

	Opt          option.Option[string]
	Und          und.Und[string]
	Elastic      elastic.Elastic[string]
	SliceUnd     sliceund.Und[string]
	SliceElastic sliceelastic.Elastic[string]
}

//undgen:ignore
type Ignored struct {
	Foo string
	Bar int
}

type Hmm struct {
	Ah Ignored
}
