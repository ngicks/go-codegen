package targettypes

import (
	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub"
)

//undgen:generated
type IncludesImplementorPlain struct {
	Foo sub.FooPlain[string]
}

//undgen:generated
type NestedImplementor2Plain struct {
	Foo sub.IncludesImplementorPlain
}
