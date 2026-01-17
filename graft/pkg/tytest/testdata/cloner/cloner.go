package cloner

type C struct{}

func (c C) Clone() C {
	return C{}
}

type CP struct{}

func (c *CP) Clone() CP {
	return CP{}
}

type Param[T, U any] struct{}

func (p Param[T, U]) CloneFunc(cloneT func(T) T, cloneU func(U) U) Param[T, U] {
	return Param[T, U]{}
}

type Total[T, U any] struct {
	C   C
	CP  CP
	CP2 *CP
	p1  Param[T, U]
	p2  Param[string, U]
}
