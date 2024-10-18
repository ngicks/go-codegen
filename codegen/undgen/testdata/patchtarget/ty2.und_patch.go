package patchtarget

import (
	sliceund "github.com/ngicks/und"
	sliceund_1 "github.com/ngicks/und/sliceund"
)

//undgen:generated
type NameOverlappingPatch struct {
	AHHH  sliceund.Und[int]      `json:",omitzero"`
	OOOHH sliceund_1.Und[string] `json:",omitempty"`
}

//undgen:generated
func (p *NameOverlappingPatch) FromValue(v NameOverlapping) {
	//nolint
	*p = NameOverlappingPatch{
		AHHH:  v.AHHH,
		OOOHH: sliceund_1.Defined(v.OOOHH),
	}
}

//undgen:generated
func (p NameOverlappingPatch) ToValue() NameOverlapping {
	//nolint
	return NameOverlapping{
		AHHH:  p.AHHH,
		OOOHH: p.OOOHH.Value(),
	}
}

//undgen:generated
func (p NameOverlappingPatch) Merge(r NameOverlappingPatch) NameOverlappingPatch {
	//nolint
	return NameOverlappingPatch{
		AHHH:  sliceund.FromOption(r.AHHH.Unwrap().Or(p.AHHH.Unwrap())),
		OOOHH: sliceund_1.FromOption(r.OOOHH.Unwrap().Or(p.OOOHH.Unwrap())),
	}
}

//undgen:generated
func (p NameOverlappingPatch) ApplyPatch(v NameOverlapping) NameOverlapping {
	var orgP NameOverlappingPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}
