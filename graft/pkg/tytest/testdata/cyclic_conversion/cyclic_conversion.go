package cyclicconversion

type A[T any] struct{}

func (a A[T]) UndPlain() B[T] {
	return B[T]{}
}

type B[T any] struct{}

func (b B[T]) UndRaw() A[T] {
	return A[T]{}
}

type AP struct{}

func (a *AP) UndPlain() BP {
	return BP{}
}

type BP struct{}

func (b *BP) UndRaw() AP {
	return AP{}
}

type AI interface {
	UndPlain() BI
}

type BI interface {
	UndRaw() AI
}

type NotImplementor struct{}

func (n NotImplementor) UndPlain() B[any] {
	return B[any]{}
}
