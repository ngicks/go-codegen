package ignored

import "github.com/ngicks/und/option"

//codegen:ignore
type Ignored struct {
	Foo string
	Bar int
	Baz option.Option[int] `json:",omitzero" und:"required"`
}
