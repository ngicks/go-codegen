package clonedep

type OnlyExported struct {
	Foo string
}

type NoExportNoClone struct {
	foo string
}

type NoExportNoCloneNoCopyable struct {
	foo []func()
}

type Clonable struct {
	foo string
}

func (c Clonable) Clone() Clonable {
	return Clonable{
		foo: c.foo,
	}
}
