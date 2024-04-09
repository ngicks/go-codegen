package clonetarget

type Serializable struct {
	Foo string
	Bar int
	Baz bool
}

type SerializableButContainsPointer struct {
	Foo string
	Bar int
	Baz bool
	Qux *string
}

type DefinedSimple int

type DefinedNoCopy []func()

type Nested struct {
	Foo string
	Bar Serializable
	Baz SerializableButContainsPointer
	Qux uint
}

type NestedPointer struct {
	Foo string
	Bar *Serializable
	Baz *SerializableButContainsPointer
	Qux uint
}

type DeeplyNested struct {
	Foo Nested1
}

type Nested1 struct {
	Bar Nested2
}

type Nested2 struct {
	Foo string
	Bar int
	Baz bool
}
