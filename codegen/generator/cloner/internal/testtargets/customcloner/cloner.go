package customcloner

import (
	"time"
	"unique"

	"github.com/ngicks/und"
)

// cloner has built-in clone method for time.Time, [](any clone-by-assign), map[(anything)](any clone-by-assign type)

type Custom struct {
	T           time.Time
	TM          map[string]time.Time
	M           map[string]int
	B           []byte
	Implementor [][]und.Und[string]
	TypeParam   und.Und[[]string]
	H           unique.Handle[string]
}
