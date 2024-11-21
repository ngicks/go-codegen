// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen patch --help
package all

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

//codegen:generated
type AllPatch struct {
	Foo sliceund.Und[string]    `json:",omitempty"`
	Bar sliceund.Und[*string]   `json:",omitempty"`
	Baz sliceund.Und[*struct{}] `json:",omitempty"`
	Qux sliceund.Und[[]string]  `json:",omitempty"`

	UntouchedOpt      sliceund.Und[int] `json:",omitempty"`
	UntouchedUnd      und.Und[int]      `json:",omitzero"`
	UntouchedSliceUnd sliceund.Und[int] `json:",omitempty"`

	OptRequired       sliceund.Und[string] `json:"opt_required,omitempty" und:"required"`
	OptNullish        sliceund.Und[string] `json:",omitempty" und:"nullish"`
	OptDef            sliceund.Und[string] `json:",omitempty" und:"def"`
	OptNull           sliceund.Und[string] `json:",omitempty" und:"null"`
	OptUnd            sliceund.Und[string] `json:",omitempty" und:"und"`
	OptDefOrUnd       sliceund.Und[string] `json:",omitempty" und:"def,und"`
	OptDefOrNull      sliceund.Und[string] `json:",omitempty" und:"def,null"`
	OptNullOrUnd      sliceund.Und[string] `json:",omitempty" und:"null,und"`
	OptDefOrNullOrUnd sliceund.Und[string] `json:",omitempty" und:"def,null,und"`

	UndRequired       und.Und[string] `json:",omitzero" und:"required"`
	UndNullish        und.Und[string] `json:",omitzero" und:"nullish"`
	UndDef            und.Und[string] `json:",omitzero" und:"def"`
	UndNull           und.Und[string] `json:",omitzero" und:"null"`
	UndUnd            und.Und[string] `json:",omitzero" und:"und"`
	UndDefOrUnd       und.Und[string] `json:",omitzero" und:"def,und"`
	UndDefOrNull      und.Und[string] `json:",omitzero" und:"def,null"`
	UndNullOrUnd      und.Und[string] `json:",omitzero" und:"null,und"`
	UndDefOrNullOrUnd und.Und[string] `json:",omitzero" und:"def,null,und"`

	ElaRequired       elastic.Elastic[string] `json:",omitzero" und:"required"`
	ElaNullish        elastic.Elastic[string] `json:",omitzero" und:"nullish"`
	ElaDef            elastic.Elastic[string] `json:",omitzero" und:"def"`
	ElaNull           elastic.Elastic[string] `json:",omitzero" und:"null"`
	ElaUnd            elastic.Elastic[string] `json:",omitzero" und:"und"`
	ElaDefOrUnd       elastic.Elastic[string] `json:",omitzero" und:"def,und"`
	ElaDefOrNull      elastic.Elastic[string] `json:",omitzero" und:"def,null"`
	ElaNullOrUnd      elastic.Elastic[string] `json:",omitzero" und:"null,und"`
	ElaDefOrNullOrUnd elastic.Elastic[string] `json:",omitzero" und:"def,null,und"`

	ElaEqEq elastic.Elastic[string] `json:",omitzero" und:"len==1"`
	ElaGr   elastic.Elastic[string] `json:",omitzero" und:"len>1"`
	ElaGrEq elastic.Elastic[string] `json:",omitzero" und:"len>=1"`
	ElaLe   elastic.Elastic[string] `json:",omitzero" und:"len<1"`
	ElaLeEq elastic.Elastic[string] `json:",omitzero" und:"len<=1"`

	ElaEqEquRequired elastic.Elastic[string] `json:",omitzero" und:"required,len==2"`
	ElaEqEquNullish  elastic.Elastic[string] `json:",omitzero" und:"nullish,len==2"`
	ElaEqEquDef      elastic.Elastic[string] `json:",omitzero" und:"def,len==2"`
	ElaEqEquNull     elastic.Elastic[string] `json:",omitzero" und:"null,len==2"`
	ElaEqEquUnd      elastic.Elastic[string] `json:",omitzero" und:"und,len==2"`

	ElaEqEqNonNullSlice      elastic.Elastic[string] `json:",omitzero" und:"values:nonnull"`
	ElaEqEqNonNullNullSlice  elastic.Elastic[string] `json:",omitzero" und:"null,values:nonnull"`
	ElaEqEqNonNullSingle     elastic.Elastic[string] `json:",omitzero" und:"values:nonnull,len==1"`
	ElaEqEqNonNullNullSingle elastic.Elastic[string] `json:",omitzero" und:"null,values:nonnull,len==1"`
	ElaEqEqNonNull           elastic.Elastic[string] `json:",omitzero" und:"values:nonnull,len==3"`
	ElaEqEqNonNullNull       elastic.Elastic[string] `json:",omitzero" und:"null,values:nonnull,len==3"`
}

//codegen:generated
func (p *AllPatch) FromValue(v All) {
	//nolint
	*p = AllPatch{
		Foo:                      sliceund.Defined(v.Foo),
		Bar:                      sliceund.Defined(v.Bar),
		Baz:                      sliceund.Defined(v.Baz),
		Qux:                      sliceund.Defined(v.Qux),
		UntouchedOpt:             option.MapOr(v.UntouchedOpt, sliceund.Null[int](), sliceund.Defined[int]),
		UntouchedUnd:             v.UntouchedUnd,
		UntouchedSliceUnd:        v.UntouchedSliceUnd,
		OptRequired:              option.MapOr(v.OptRequired, sliceund.Null[string](), sliceund.Defined[string]),
		OptNullish:               option.MapOr(v.OptNullish, sliceund.Null[string](), sliceund.Defined[string]),
		OptDef:                   option.MapOr(v.OptDef, sliceund.Null[string](), sliceund.Defined[string]),
		OptNull:                  option.MapOr(v.OptNull, sliceund.Null[string](), sliceund.Defined[string]),
		OptUnd:                   option.MapOr(v.OptUnd, sliceund.Null[string](), sliceund.Defined[string]),
		OptDefOrUnd:              option.MapOr(v.OptDefOrUnd, sliceund.Null[string](), sliceund.Defined[string]),
		OptDefOrNull:             option.MapOr(v.OptDefOrNull, sliceund.Null[string](), sliceund.Defined[string]),
		OptNullOrUnd:             option.MapOr(v.OptNullOrUnd, sliceund.Null[string](), sliceund.Defined[string]),
		OptDefOrNullOrUnd:        option.MapOr(v.OptDefOrNullOrUnd, sliceund.Null[string](), sliceund.Defined[string]),
		UndRequired:              v.UndRequired,
		UndNullish:               v.UndNullish,
		UndDef:                   v.UndDef,
		UndNull:                  v.UndNull,
		UndUnd:                   v.UndUnd,
		UndDefOrUnd:              v.UndDefOrUnd,
		UndDefOrNull:             v.UndDefOrNull,
		UndNullOrUnd:             v.UndNullOrUnd,
		UndDefOrNullOrUnd:        v.UndDefOrNullOrUnd,
		ElaRequired:              v.ElaRequired,
		ElaNullish:               v.ElaNullish,
		ElaDef:                   v.ElaDef,
		ElaNull:                  v.ElaNull,
		ElaUnd:                   v.ElaUnd,
		ElaDefOrUnd:              v.ElaDefOrUnd,
		ElaDefOrNull:             v.ElaDefOrNull,
		ElaNullOrUnd:             v.ElaNullOrUnd,
		ElaDefOrNullOrUnd:        v.ElaDefOrNullOrUnd,
		ElaEqEq:                  v.ElaEqEq,
		ElaGr:                    v.ElaGr,
		ElaGrEq:                  v.ElaGrEq,
		ElaLe:                    v.ElaLe,
		ElaLeEq:                  v.ElaLeEq,
		ElaEqEquRequired:         v.ElaEqEquRequired,
		ElaEqEquNullish:          v.ElaEqEquNullish,
		ElaEqEquDef:              v.ElaEqEquDef,
		ElaEqEquNull:             v.ElaEqEquNull,
		ElaEqEquUnd:              v.ElaEqEquUnd,
		ElaEqEqNonNullSlice:      v.ElaEqEqNonNullSlice,
		ElaEqEqNonNullNullSlice:  v.ElaEqEqNonNullNullSlice,
		ElaEqEqNonNullSingle:     v.ElaEqEqNonNullSingle,
		ElaEqEqNonNullNullSingle: v.ElaEqEqNonNullNullSingle,
		ElaEqEqNonNull:           v.ElaEqEqNonNull,
		ElaEqEqNonNullNull:       v.ElaEqEqNonNullNull,
	}
}

//codegen:generated
func (p AllPatch) ToValue() All {
	//nolint
	return All{
		Foo:                      p.Foo.Value(),
		Bar:                      p.Bar.Value(),
		Baz:                      p.Baz.Value(),
		Qux:                      p.Qux.Value(),
		UntouchedOpt:             option.Flatten(p.UntouchedOpt.Unwrap()),
		UntouchedUnd:             p.UntouchedUnd,
		UntouchedSliceUnd:        p.UntouchedSliceUnd,
		OptRequired:              option.Flatten(p.OptRequired.Unwrap()),
		OptNullish:               option.Flatten(p.OptNullish.Unwrap()),
		OptDef:                   option.Flatten(p.OptDef.Unwrap()),
		OptNull:                  option.Flatten(p.OptNull.Unwrap()),
		OptUnd:                   option.Flatten(p.OptUnd.Unwrap()),
		OptDefOrUnd:              option.Flatten(p.OptDefOrUnd.Unwrap()),
		OptDefOrNull:             option.Flatten(p.OptDefOrNull.Unwrap()),
		OptNullOrUnd:             option.Flatten(p.OptNullOrUnd.Unwrap()),
		OptDefOrNullOrUnd:        option.Flatten(p.OptDefOrNullOrUnd.Unwrap()),
		UndRequired:              p.UndRequired,
		UndNullish:               p.UndNullish,
		UndDef:                   p.UndDef,
		UndNull:                  p.UndNull,
		UndUnd:                   p.UndUnd,
		UndDefOrUnd:              p.UndDefOrUnd,
		UndDefOrNull:             p.UndDefOrNull,
		UndNullOrUnd:             p.UndNullOrUnd,
		UndDefOrNullOrUnd:        p.UndDefOrNullOrUnd,
		ElaRequired:              p.ElaRequired,
		ElaNullish:               p.ElaNullish,
		ElaDef:                   p.ElaDef,
		ElaNull:                  p.ElaNull,
		ElaUnd:                   p.ElaUnd,
		ElaDefOrUnd:              p.ElaDefOrUnd,
		ElaDefOrNull:             p.ElaDefOrNull,
		ElaNullOrUnd:             p.ElaNullOrUnd,
		ElaDefOrNullOrUnd:        p.ElaDefOrNullOrUnd,
		ElaEqEq:                  p.ElaEqEq,
		ElaGr:                    p.ElaGr,
		ElaGrEq:                  p.ElaGrEq,
		ElaLe:                    p.ElaLe,
		ElaLeEq:                  p.ElaLeEq,
		ElaEqEquRequired:         p.ElaEqEquRequired,
		ElaEqEquNullish:          p.ElaEqEquNullish,
		ElaEqEquDef:              p.ElaEqEquDef,
		ElaEqEquNull:             p.ElaEqEquNull,
		ElaEqEquUnd:              p.ElaEqEquUnd,
		ElaEqEqNonNullSlice:      p.ElaEqEqNonNullSlice,
		ElaEqEqNonNullNullSlice:  p.ElaEqEqNonNullNullSlice,
		ElaEqEqNonNullSingle:     p.ElaEqEqNonNullSingle,
		ElaEqEqNonNullNullSingle: p.ElaEqEqNonNullNullSingle,
		ElaEqEqNonNull:           p.ElaEqEqNonNull,
		ElaEqEqNonNullNull:       p.ElaEqEqNonNullNull,
	}
}

//codegen:generated
func (p AllPatch) Merge(r AllPatch) AllPatch {
	//nolint
	return AllPatch{
		Foo:                      sliceund.FromOption(r.Foo.Unwrap().Or(p.Foo.Unwrap())),
		Bar:                      sliceund.FromOption(r.Bar.Unwrap().Or(p.Bar.Unwrap())),
		Baz:                      sliceund.FromOption(r.Baz.Unwrap().Or(p.Baz.Unwrap())),
		Qux:                      sliceund.FromOption(r.Qux.Unwrap().Or(p.Qux.Unwrap())),
		UntouchedOpt:             sliceund.FromOption(r.UntouchedOpt.Unwrap().Or(p.UntouchedOpt.Unwrap())),
		UntouchedUnd:             und.FromOption(r.UntouchedUnd.Unwrap().Or(p.UntouchedUnd.Unwrap())),
		UntouchedSliceUnd:        sliceund.FromOption(r.UntouchedSliceUnd.Unwrap().Or(p.UntouchedSliceUnd.Unwrap())),
		OptRequired:              sliceund.FromOption(r.OptRequired.Unwrap().Or(p.OptRequired.Unwrap())),
		OptNullish:               sliceund.FromOption(r.OptNullish.Unwrap().Or(p.OptNullish.Unwrap())),
		OptDef:                   sliceund.FromOption(r.OptDef.Unwrap().Or(p.OptDef.Unwrap())),
		OptNull:                  sliceund.FromOption(r.OptNull.Unwrap().Or(p.OptNull.Unwrap())),
		OptUnd:                   sliceund.FromOption(r.OptUnd.Unwrap().Or(p.OptUnd.Unwrap())),
		OptDefOrUnd:              sliceund.FromOption(r.OptDefOrUnd.Unwrap().Or(p.OptDefOrUnd.Unwrap())),
		OptDefOrNull:             sliceund.FromOption(r.OptDefOrNull.Unwrap().Or(p.OptDefOrNull.Unwrap())),
		OptNullOrUnd:             sliceund.FromOption(r.OptNullOrUnd.Unwrap().Or(p.OptNullOrUnd.Unwrap())),
		OptDefOrNullOrUnd:        sliceund.FromOption(r.OptDefOrNullOrUnd.Unwrap().Or(p.OptDefOrNullOrUnd.Unwrap())),
		UndRequired:              und.FromOption(r.UndRequired.Unwrap().Or(p.UndRequired.Unwrap())),
		UndNullish:               und.FromOption(r.UndNullish.Unwrap().Or(p.UndNullish.Unwrap())),
		UndDef:                   und.FromOption(r.UndDef.Unwrap().Or(p.UndDef.Unwrap())),
		UndNull:                  und.FromOption(r.UndNull.Unwrap().Or(p.UndNull.Unwrap())),
		UndUnd:                   und.FromOption(r.UndUnd.Unwrap().Or(p.UndUnd.Unwrap())),
		UndDefOrUnd:              und.FromOption(r.UndDefOrUnd.Unwrap().Or(p.UndDefOrUnd.Unwrap())),
		UndDefOrNull:             und.FromOption(r.UndDefOrNull.Unwrap().Or(p.UndDefOrNull.Unwrap())),
		UndNullOrUnd:             und.FromOption(r.UndNullOrUnd.Unwrap().Or(p.UndNullOrUnd.Unwrap())),
		UndDefOrNullOrUnd:        und.FromOption(r.UndDefOrNullOrUnd.Unwrap().Or(p.UndDefOrNullOrUnd.Unwrap())),
		ElaRequired:              elastic.FromUnd(und.FromOption(r.ElaRequired.Unwrap().Unwrap().Or(p.ElaRequired.Unwrap().Unwrap()))),
		ElaNullish:               elastic.FromUnd(und.FromOption(r.ElaNullish.Unwrap().Unwrap().Or(p.ElaNullish.Unwrap().Unwrap()))),
		ElaDef:                   elastic.FromUnd(und.FromOption(r.ElaDef.Unwrap().Unwrap().Or(p.ElaDef.Unwrap().Unwrap()))),
		ElaNull:                  elastic.FromUnd(und.FromOption(r.ElaNull.Unwrap().Unwrap().Or(p.ElaNull.Unwrap().Unwrap()))),
		ElaUnd:                   elastic.FromUnd(und.FromOption(r.ElaUnd.Unwrap().Unwrap().Or(p.ElaUnd.Unwrap().Unwrap()))),
		ElaDefOrUnd:              elastic.FromUnd(und.FromOption(r.ElaDefOrUnd.Unwrap().Unwrap().Or(p.ElaDefOrUnd.Unwrap().Unwrap()))),
		ElaDefOrNull:             elastic.FromUnd(und.FromOption(r.ElaDefOrNull.Unwrap().Unwrap().Or(p.ElaDefOrNull.Unwrap().Unwrap()))),
		ElaNullOrUnd:             elastic.FromUnd(und.FromOption(r.ElaNullOrUnd.Unwrap().Unwrap().Or(p.ElaNullOrUnd.Unwrap().Unwrap()))),
		ElaDefOrNullOrUnd:        elastic.FromUnd(und.FromOption(r.ElaDefOrNullOrUnd.Unwrap().Unwrap().Or(p.ElaDefOrNullOrUnd.Unwrap().Unwrap()))),
		ElaEqEq:                  elastic.FromUnd(und.FromOption(r.ElaEqEq.Unwrap().Unwrap().Or(p.ElaEqEq.Unwrap().Unwrap()))),
		ElaGr:                    elastic.FromUnd(und.FromOption(r.ElaGr.Unwrap().Unwrap().Or(p.ElaGr.Unwrap().Unwrap()))),
		ElaGrEq:                  elastic.FromUnd(und.FromOption(r.ElaGrEq.Unwrap().Unwrap().Or(p.ElaGrEq.Unwrap().Unwrap()))),
		ElaLe:                    elastic.FromUnd(und.FromOption(r.ElaLe.Unwrap().Unwrap().Or(p.ElaLe.Unwrap().Unwrap()))),
		ElaLeEq:                  elastic.FromUnd(und.FromOption(r.ElaLeEq.Unwrap().Unwrap().Or(p.ElaLeEq.Unwrap().Unwrap()))),
		ElaEqEquRequired:         elastic.FromUnd(und.FromOption(r.ElaEqEquRequired.Unwrap().Unwrap().Or(p.ElaEqEquRequired.Unwrap().Unwrap()))),
		ElaEqEquNullish:          elastic.FromUnd(und.FromOption(r.ElaEqEquNullish.Unwrap().Unwrap().Or(p.ElaEqEquNullish.Unwrap().Unwrap()))),
		ElaEqEquDef:              elastic.FromUnd(und.FromOption(r.ElaEqEquDef.Unwrap().Unwrap().Or(p.ElaEqEquDef.Unwrap().Unwrap()))),
		ElaEqEquNull:             elastic.FromUnd(und.FromOption(r.ElaEqEquNull.Unwrap().Unwrap().Or(p.ElaEqEquNull.Unwrap().Unwrap()))),
		ElaEqEquUnd:              elastic.FromUnd(und.FromOption(r.ElaEqEquUnd.Unwrap().Unwrap().Or(p.ElaEqEquUnd.Unwrap().Unwrap()))),
		ElaEqEqNonNullSlice:      elastic.FromUnd(und.FromOption(r.ElaEqEqNonNullSlice.Unwrap().Unwrap().Or(p.ElaEqEqNonNullSlice.Unwrap().Unwrap()))),
		ElaEqEqNonNullNullSlice:  elastic.FromUnd(und.FromOption(r.ElaEqEqNonNullNullSlice.Unwrap().Unwrap().Or(p.ElaEqEqNonNullNullSlice.Unwrap().Unwrap()))),
		ElaEqEqNonNullSingle:     elastic.FromUnd(und.FromOption(r.ElaEqEqNonNullSingle.Unwrap().Unwrap().Or(p.ElaEqEqNonNullSingle.Unwrap().Unwrap()))),
		ElaEqEqNonNullNullSingle: elastic.FromUnd(und.FromOption(r.ElaEqEqNonNullNullSingle.Unwrap().Unwrap().Or(p.ElaEqEqNonNullNullSingle.Unwrap().Unwrap()))),
		ElaEqEqNonNull:           elastic.FromUnd(und.FromOption(r.ElaEqEqNonNull.Unwrap().Unwrap().Or(p.ElaEqEqNonNull.Unwrap().Unwrap()))),
		ElaEqEqNonNullNull:       elastic.FromUnd(und.FromOption(r.ElaEqEqNonNullNull.Unwrap().Unwrap().Or(p.ElaEqEqNonNullNull.Unwrap().Unwrap()))),
	}
}

//codegen:generated
func (p AllPatch) ApplyPatch(v All) All {
	var orgP AllPatch
	orgP.FromValue(v)
	merged := orgP.Merge(p)
	return merged.ToValue()
}
