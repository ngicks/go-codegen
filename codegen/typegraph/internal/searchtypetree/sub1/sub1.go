package sub1

import (
	"github.com/ngicks/go-codegen/codegen/typegraph/internal/searchtypetree/sub2"
	"github.com/ngicks/und/option"
)

// matched type.
type Foo struct {
	O option.Option[string] `und:"def"`
}

// nested
type Bar struct {
	O [][]map[string]option.Option[string] `und:"def"`
}

type HasAlias struct {
	F Baz
}

// alias to local implementor
type Baz = Bar

type HasAliasToImplementor struct {
	Qux Qux
}

// alias to implementor.
type Qux = sub2.Foo
