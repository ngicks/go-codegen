package sub

import (
	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub2"
)

//undgen:generated
type IncludesImplementorPlain struct {
	Foo sub2.FooPlain[int]
}

func (v IncludesImplementor) UndPlain() IncludesImplementorPlain {
	return IncludesImplementorPlain{
		Foo: v.Foo.UndPlain(),
	}
}

func (v IncludesImplementorPlain) UndRaw() IncludesImplementor {
	return IncludesImplementor{
		Foo: v.Foo.UndRaw(),
	}
}
