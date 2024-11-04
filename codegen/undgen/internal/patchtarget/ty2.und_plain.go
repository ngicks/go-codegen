package patchtarget

import (
	sliceund "github.com/ngicks/und"
)

//undgen:generated
type NameOverlappingPlain struct {
	AHHH  int `und:"required"`
	OOOHH string
}

func (v NameOverlapping) UndPlain() NameOverlappingPlain {
	return NameOverlappingPlain{
		AHHH:  v.AHHH.Value(),
		OOOHH: v.OOOHH,
	}
}

func (v NameOverlappingPlain) UndRaw() NameOverlapping {
	return NameOverlapping{
		AHHH:  sliceund.Defined(v.AHHH),
		OOOHH: v.OOOHH,
	}
}
