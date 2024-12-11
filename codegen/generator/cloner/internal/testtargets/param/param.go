package param

type Param[T, U any] struct {
	U U
	T T
}

type Param2[T, U any] struct {
	U map[string]*U
	T *T
}
