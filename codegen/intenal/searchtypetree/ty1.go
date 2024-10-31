package searchtypetree

import (
	"github.com/ngicks/und/option"
)

type Foo struct {
	O option.Option[string] `und:"def"`
}

type Bar struct {
	O []option.Option[string] `und:"def"`
}

type Baz struct {
	O TypeParam[option.Option[string]]
}

type TypeParam[T any] struct {
	T T
}
