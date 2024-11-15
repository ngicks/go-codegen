package typeparam

import (
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
)

type WithTypeParam[T any] struct {
	Foo string
	Bar T
	Baz option.Option[T]              `json:",omitzero" und:"required"`
	Qux map[string]elastic.Elastic[T] `json:"qux" und:"len==2,values:nonnull"`
}
