package sub2

//undgen:ignore
type Foo[T any] struct {
	T   T
	Yay string
}

func (f Foo[T]) UndPlain() FooPlain[T] {
	return FooPlain[T]{
		Nay: f.Yay,
	}
}

//undgen:ignore
type FooPlain[T any] struct {
	T   T
	Nay string
}

func (f FooPlain[T]) UndRaw() Foo[T] {
	return Foo[T]{
		Yay: f.Nay,
	}
}
