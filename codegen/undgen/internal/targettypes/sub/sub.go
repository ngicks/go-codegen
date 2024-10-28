package sub

import (
	"github.com/ngicks/go-codegen/codegen/undgen/internal/targettypes/sub2"
	"github.com/ngicks/und/option"
)

//undgen:ignore
type Foo[T any] struct {
	T   T
	Yay string
}

func (f Foo[T]) UndPlain() FooPlain[T] {
	return FooPlain[T]{
		T: f.T,
		Nay: func() string {
			if f.Yay == "yay" {
				return "nay"
			} else {
				return f.Yay
			}
		}(),
	}
}

//undgen:ignore
type FooPlain[T any] struct {
	T   T
	Nay string
}

func (f FooPlain[T]) UndRaw() Foo[T] {
	return Foo[T]{
		T: f.T,
		Yay: func() string {
			if f.Nay == "nay" {
				return "yay"
			} else {
				return f.Nay
			}
		}(),
	}
}

//undgen:ignore
type Bar struct {
	O option.Option[string]
}

type Baz[T any] struct {
	O option.Option[T]
}

//undgen:ignore
type NonCyclic struct {
	Foo string
}

func (nc NonCyclic) UndRaw() struct{} {
	return struct{}{}
}

type IncludesImplementor struct {
	Foo sub2.Foo[int]
}
