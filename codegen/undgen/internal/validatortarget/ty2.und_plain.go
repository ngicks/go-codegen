package validatortarget

import (
	"github.com/ngicks/und/option"
)

//undgen:generated
type CPlain [3]option.Option[AllPlain]

//undgen:generated
type DPlain struct {
	Foo AllPlain
	Bar AllPlain `und:"required"`
}
