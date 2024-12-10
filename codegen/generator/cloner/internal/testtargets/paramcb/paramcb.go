package paramcb

type A[T any] struct {
	A B[string, T]
	B C[[]string]
	C B[C[string], []C[string]]
}

type B[T, U any] struct {
	T T
	U U
}

type C[T any] struct {
	T T
}
