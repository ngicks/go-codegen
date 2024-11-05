package patchtarget

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
