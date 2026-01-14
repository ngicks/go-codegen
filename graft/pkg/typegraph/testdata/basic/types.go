package basic

type (
	TypeA struct {
		B TypeB
		C *TypeC
	}

	TypeB struct {
		Name string
	}

	TypeC struct {
		D TypeD
	}

	TypeD struct {
		Value int
	}
)
