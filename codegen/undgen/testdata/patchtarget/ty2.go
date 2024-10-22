package patchtarget

import (
	// intentionally overlapping name to sliceund
	// to see if it can handle name overlap correctly
	sliceund "github.com/ngicks/und"
)

type NameOverlapping struct {
	AHHH  sliceund.Und[int] `und:"required"`
	OOOHH string
}
