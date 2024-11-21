// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen patch --help
package deeplynested

import (
	"github.com/ngicks/go-codegen/codegen/generator/undgen/internal/testtargets/implementor"
	"github.com/ngicks/und"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

//codegen:generated
type DeeplyNestedImplementorPatch struct {
	A sliceund.Und[[]map[string][5]und.Und[implementor.Implementor[string]]]  `und:"required" json:",omitempty"`
	B sliceund.Und[[][][]map[int]implementor.Implementor[string]]             `json:",omitempty"`
	C sliceund.Und[[]map[string][5]und.Und[*implementor.Implementor[string]]] `und:"required" json:",omitempty"`
	D sliceund.Und[[][][]map[int]*implementor.Implementor[string]]            `json:",omitempty"`
}

//codegen:generated
func (p *DeeplyNestedImplementorPatch) FromValue(v DeeplyNestedImplementor) {
	//nolint
	*p = DeeplyNestedImplementorPatch{
		A: sliceund.Defined(v.A),
		B: sliceund.Defined(v.B),
		C: sliceund.Defined(v.C),
		D: sliceund.Defined(v.D),
	}
}

//codegen:generated
func (p DeeplyNestedImplementorPatch) ToValue() DeeplyNestedImplementor {
	//nolint
	return DeeplyNestedImplementor{
		A: p.A.Value(),
		B: p.B.Value(),
		C: p.C.Value(),
		D: p.D.Value(),
	}
}

//codegen:generated
func (p DeeplyNestedImplementorPatch) Merge(r DeeplyNestedImplementorPatch) DeeplyNestedImplementorPatch {
	//nolint
	return DeeplyNestedImplementorPatch{
		A: sliceund.FromOption(r.A.Unwrap().Or(p.A.Unwrap())),
		B: sliceund.FromOption(r.B.Unwrap().Or(p.B.Unwrap())),
		C: sliceund.FromOption(r.C.Unwrap().Or(p.C.Unwrap())),
		D: sliceund.FromOption(r.D.Unwrap().Or(p.D.Unwrap())),
	}
}

//codegen:generated
func (p DeeplyNestedImplementorPatch) ApplyPatch(v DeeplyNestedImplementor) DeeplyNestedImplementor {
	var orgP DeeplyNestedImplementorPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}

//codegen:generated
type DependantPatch struct {
	Opt sliceund.Und[string] `und:"required" json:",omitempty"`
}

//codegen:generated
func (p *DependantPatch) FromValue(v Dependant) {
	//nolint
	*p = DependantPatch{
		Opt: option.MapOr(v.Opt, sliceund.Null[string](), sliceund.Defined[string]),
	}
}

//codegen:generated
func (p DependantPatch) ToValue() Dependant {
	//nolint
	return Dependant{
		Opt: option.Flatten(p.Opt.Unwrap()),
	}
}

//codegen:generated
func (p DependantPatch) Merge(r DependantPatch) DependantPatch {
	//nolint
	return DependantPatch{
		Opt: sliceund.FromOption(r.Opt.Unwrap().Or(p.Opt.Unwrap())),
	}
}

//codegen:generated
func (p DependantPatch) ApplyPatch(v Dependant) Dependant {
	var orgP DependantPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}

//codegen:generated
type DeeplyNestedDependantPatch struct {
	A sliceund.Und[[]map[string][5]und.Und[Dependant]]  `und:"required" json:",omitempty"`
	B sliceund.Und[[][][]map[int]Dependant]             `json:",omitempty"`
	C sliceund.Und[[]map[string][5]und.Und[*Dependant]] `und:"required" json:",omitempty"`
	D sliceund.Und[[][][]map[int]*Dependant]            `json:",omitempty"`
}

//codegen:generated
func (p *DeeplyNestedDependantPatch) FromValue(v DeeplyNestedDependant) {
	//nolint
	*p = DeeplyNestedDependantPatch{
		A: sliceund.Defined(v.A),
		B: sliceund.Defined(v.B),
		C: sliceund.Defined(v.C),
		D: sliceund.Defined(v.D),
	}
}

//codegen:generated
func (p DeeplyNestedDependantPatch) ToValue() DeeplyNestedDependant {
	//nolint
	return DeeplyNestedDependant{
		A: p.A.Value(),
		B: p.B.Value(),
		C: p.C.Value(),
		D: p.D.Value(),
	}
}

//codegen:generated
func (p DeeplyNestedDependantPatch) Merge(r DeeplyNestedDependantPatch) DeeplyNestedDependantPatch {
	//nolint
	return DeeplyNestedDependantPatch{
		A: sliceund.FromOption(r.A.Unwrap().Or(p.A.Unwrap())),
		B: sliceund.FromOption(r.B.Unwrap().Or(p.B.Unwrap())),
		C: sliceund.FromOption(r.C.Unwrap().Or(p.C.Unwrap())),
		D: sliceund.FromOption(r.D.Unwrap().Or(p.D.Unwrap())),
	}
}

//codegen:generated
func (p DeeplyNestedDependantPatch) ApplyPatch(v DeeplyNestedDependant) DeeplyNestedDependant {
	var orgP DeeplyNestedDependantPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}

//codegen:generated
type DeeplyNestedImplementorMapPatch []map[string][5]und.Und[implementor.Implementor[string]]

//codegen:generated
func (p *DeeplyNestedImplementorMapPatch) FromValue(v DeeplyNestedImplementorMap) {
	//nolint
	*p = DeeplyNestedImplementorMapPatch{}
}

//codegen:generated
func (p DeeplyNestedImplementorMapPatch) ToValue() DeeplyNestedImplementorMap {
	//nolint
	return DeeplyNestedImplementorMap{}
}

//codegen:generated
func (p DeeplyNestedImplementorMapPatch) Merge(r DeeplyNestedImplementorMapPatch) DeeplyNestedImplementorMapPatch {
	//nolint
	return DeeplyNestedImplementorMapPatch{}
}

//codegen:generated
func (p DeeplyNestedImplementorMapPatch) ApplyPatch(v DeeplyNestedImplementorMap) DeeplyNestedImplementorMap {
	var orgP DeeplyNestedImplementorMapPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}

//codegen:generated
type DeeplyNestedDependantMapPatch []map[string][5]und.Und[Dependant]

//codegen:generated
func (p *DeeplyNestedDependantMapPatch) FromValue(v DeeplyNestedDependantMap) {
	//nolint
	*p = DeeplyNestedDependantMapPatch{}
}

//codegen:generated
func (p DeeplyNestedDependantMapPatch) ToValue() DeeplyNestedDependantMap {
	//nolint
	return DeeplyNestedDependantMap{}
}

//codegen:generated
func (p DeeplyNestedDependantMapPatch) Merge(r DeeplyNestedDependantMapPatch) DeeplyNestedDependantMapPatch {
	//nolint
	return DeeplyNestedDependantMapPatch{}
}

//codegen:generated
func (p DeeplyNestedDependantMapPatch) ApplyPatch(v DeeplyNestedDependantMap) DeeplyNestedDependantMap {
	var orgP DeeplyNestedDependantMapPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}
