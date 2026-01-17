package error

type A struct{}

func (a A) A() error {
	return nil
}

func (a *A) B() error {
	return nil
}

func (a A) C() struct{} {
	return struct{}{}
}

func (a *A) D() struct{} {
	return struct{}{}
}
