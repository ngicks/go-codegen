package disalloweddependant

type DisallowedDependant struct {
	Chan chan int
	A    A
}

type A struct {
	Chan chan int
	B    B
}

type B struct {
	B int
}
