package alias

import "github.com/ngicks/und"

type U = und.Und[string]

type U2 = []und.Und[string]

type U3 = map[string]U2

type U4 = U

type U5 = U4

type A struct {
	U  U
	UM map[string]U
}

type B struct {
	U2 U2
}

type C struct {
	U3 U3
}

type D struct {
	U5 U5
}
