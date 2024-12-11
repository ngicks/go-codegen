package alias

import "github.com/ngicks/und"

type U = und.Und[string]

type A struct {
	U  U
	UM map[string]U
}
