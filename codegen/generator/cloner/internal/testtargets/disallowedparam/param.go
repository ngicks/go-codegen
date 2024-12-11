package disallowedparam

import (
	"sync"

	"github.com/ngicks/und"
)

type A struct {
	U       und.Und[string]
	Ignored und.Und[*sync.Mutex]
}

type B struct {
	U          und.Und[string]
	Disallowed und.Und[chan int]
}
