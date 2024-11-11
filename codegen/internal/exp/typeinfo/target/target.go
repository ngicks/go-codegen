package target

type Foo struct {
	Bar func()
}

func (f Foo) MethodOnNonPointer() {
	//
}

func (f *Foo) MethodOnPointer() {
	//
}
