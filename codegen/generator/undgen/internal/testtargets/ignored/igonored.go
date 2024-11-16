package ignored

import "github.com/ngicks/und/option"

//undgen:ignore
type Ignored struct {
	Foo string
	Bar int
	Baz option.Option[int] `json:",omitzero" und:"required"`
}
