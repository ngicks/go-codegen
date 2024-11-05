package targettypes

import (
	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub"
)

//undgen:generated
type IncludesImplementorPlain struct {
	Foo sub.FooPlain[string]
}

func (v IncludesImplementor) UndPlain() IncludesImplementorPlain {
	return IncludesImplementorPlain{
		Foo: v.Foo.UndPlain(),
	}
}

//undgen:generated
type NestedImplementor2Plain struct {
	Foo sub.IncludesImplementorPlain
}

func (v NestedImplementor2) UndPlain() NestedImplementor2Plain {
	return NestedImplementor2Plain{
		Foo: v.Foo.UndPlain(),
	}
}
