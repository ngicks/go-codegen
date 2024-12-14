package structlit

type A struct {
	A int
	B struct {
		Foo string
		Bar *[]struct {
			Baz B
			Qux string
		}
	}
}

type B struct {
	B string
}

type C struct {
	Foo struct {
		Bar struct {
			Baz string
			Qux int
		}
		Quux float64
	}
}
