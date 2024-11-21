package implementor

import "fmt"

//codegen:ignore
type Implementor[T any] struct {
	T   T
	Yay string
}

func (f Implementor[T]) UndValidate() error {
	if f.Yay == "nay" {
		return fmt.Errorf("you said \"nay\"?????????")
	}
	return nil
}

//codegen:ignore
func (f Implementor[T]) UndPlain() ImplementorPlain[T] {
	return ImplementorPlain[T]{
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

//codegen:ignore
type ImplementorPlain[T any] struct {
	T   T
	Nay string
}

func (f ImplementorPlain[T]) UndRaw() Implementor[T] {
	return Implementor[T]{
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
