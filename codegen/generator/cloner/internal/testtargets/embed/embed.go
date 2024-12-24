package embed

import "slices"

type A struct {
	Embed
}

type B struct {
	F1 string
	Embed
	F2 int
}

type C struct {
	EmbedImplementor
}

type D struct {
	F1 string
	EmbedImplementor
	F2 int
}

//codegen:ignore
type Embed struct {
	Foo string
	Bar []int
}

//codegen:ignore
type EmbedImplementor struct {
	foo string
	bar []int
}

func (e EmbedImplementor) Clone() EmbedImplementor {
	return EmbedImplementor{
		foo: e.foo,
		bar: slices.Clone(e.bar),
	}
}
