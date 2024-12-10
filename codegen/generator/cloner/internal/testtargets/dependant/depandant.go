package dependant

type Root struct {
	A A
}

type A struct {
	B B
	C C
}

type B struct {
	C C
}

type C struct {
	C string
}
