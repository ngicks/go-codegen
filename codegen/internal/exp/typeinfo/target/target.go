package target

type Foo struct {
}

func (f Foo) MethodOnNonPointer() {
	//
}

func (f *Foo) MethodOnPointer() {
	//
}
