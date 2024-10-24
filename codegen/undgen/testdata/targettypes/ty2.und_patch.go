package targettypes

import (
	"github.com/ngicks/go-codegen/codegen/undgen/testdata/targettypes/sub"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

//undgen:generated
type APatch struct {
	A sliceund.Und[string] `json:",omitempty"`
}

//undgen:generated
func (p *APatch) FromValue(v A) {
	//nolint
	*p = APatch{
		A: option.MapOrOption(v.A, sliceund.Null[string](), sliceund.Defined[string]),
	}
}

//undgen:generated
func (p APatch) ToValue() A {
	//nolint
	return A{
		A: option.FlattenOption(p.A.Unwrap()),
	}
}

//undgen:generated
func (p APatch) Merge(r APatch) APatch {
	//nolint
	return APatch{
		A: sliceund.FromOption(r.A.Unwrap().Or(p.A.Unwrap())),
	}
}

//undgen:generated
func (p APatch) ApplyPatch(v A) A {
	var orgP APatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}

//undgen:generated
type BPatch struct {
	B und.Und[int] `json:",omitzero"`
}

//undgen:generated
func (p *BPatch) FromValue(v B) {
	//nolint
	*p = BPatch{
		B: v.B,
	}
}

//undgen:generated
func (p BPatch) ToValue() B {
	//nolint
	return B{
		B: p.B,
	}
}

//undgen:generated
func (p BPatch) Merge(r BPatch) BPatch {
	//nolint
	return BPatch{
		B: und.FromOption(r.B.Unwrap().Or(p.B.Unwrap())),
	}
}

//undgen:generated
func (p BPatch) ApplyPatch(v B) B {
	var orgP BPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}

//undgen:generated
type IncludesSubTargetPatch struct {
	Foo sliceund.Und[sub.Baz[string]] `json:",omitempty"`
}

//undgen:generated
func (p *IncludesSubTargetPatch) FromValue(v IncludesSubTarget) {
	//nolint
	*p = IncludesSubTargetPatch{
		Foo: sliceund.Defined(v.Foo),
	}
}

//undgen:generated
func (p IncludesSubTargetPatch) ToValue() IncludesSubTarget {
	//nolint
	return IncludesSubTarget{
		Foo: p.Foo.Value(),
	}
}

//undgen:generated
func (p IncludesSubTargetPatch) Merge(r IncludesSubTargetPatch) IncludesSubTargetPatch {
	//nolint
	return IncludesSubTargetPatch{
		Foo: sliceund.FromOption(r.Foo.Unwrap().Or(p.Foo.Unwrap())),
	}
}

//undgen:generated
func (p IncludesSubTargetPatch) ApplyPatch(v IncludesSubTarget) IncludesSubTarget {
	var orgP IncludesSubTargetPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}
