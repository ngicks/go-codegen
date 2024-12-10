package implementor

import "github.com/ngicks/und"

type ContainsImplementor[T any] struct {
	U  und.Und[T]
	US und.Und[string]
	W  Wow
}

//codegen:ignore
type Wow und.Und[string]

func (w Wow) Clone() Wow {
	return Wow(und.Und[string](w).CloneFunc(func(s string) string { return s }))
}
