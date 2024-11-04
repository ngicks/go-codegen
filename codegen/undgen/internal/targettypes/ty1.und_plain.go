package targettypes

import (
	"github.com/ngicks/und"
	"github.com/ngicks/und/conversion"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	"github.com/ngicks/und/sliceund"
)

//undgen:generated
type AllPlain struct {
	Foo string
	Bar *string
	Baz *struct{}
	Qux []string

	UntouchedOpt      option.Option[int] `json:",omitzero"`
	UntouchedUnd      und.Und[int]       `json:",omitzero"`
	UntouchedSliceUnd sliceund.Und[int]  `json:",omitzero"`

	OptRequired       string                `json:"opt_required,omitzero" und:"required"`
	OptNullish        conversion.Empty      `json:",omitzero" und:"nullish"`
	OptDef            string                `json:",omitzero" und:"def"`
	OptNull           conversion.Empty      `json:",omitzero" und:"null"`
	OptUnd            conversion.Empty      `json:",omitzero" und:"und"`
	OptDefOrUnd       option.Option[string] `json:",omitzero" und:"def,und"`
	OptDefOrNull      option.Option[string] `json:",omitzero" und:"def,null"`
	OptNullOrUnd      conversion.Empty      `json:",omitzero" und:"null,und"`
	OptDefOrNullOrUnd option.Option[string] `json:",omitzero" und:"def,null,und"`

	UndRequired       string                          `json:",omitzero" und:"required"`
	UndNullish        option.Option[conversion.Empty] `json:",omitzero" und:"nullish"`
	UndDef            string                          `json:",omitzero" und:"def"`
	UndNull           conversion.Empty                `json:",omitzero" und:"null"`
	UndUnd            conversion.Empty                `json:",omitzero" und:"und"`
	UndDefOrUnd       option.Option[string]           `json:",omitzero" und:"def,und"`
	UndDefOrNull      option.Option[string]           `json:",omitzero" und:"def,null"`
	UndNullOrUnd      option.Option[conversion.Empty] `json:",omitzero" und:"null,und"`
	UndDefOrNullOrUnd und.Und[string]                 `json:",omitzero" und:"def,null,und"`

	ElaRequired       []option.Option[string]                `json:",omitzero" und:"required"`
	ElaNullish        option.Option[conversion.Empty]        `json:",omitzero" und:"nullish"`
	ElaDef            []option.Option[string]                `json:",omitzero" und:"def"`
	ElaNull           conversion.Empty                       `json:",omitzero" und:"null"`
	ElaUnd            conversion.Empty                       `json:",omitzero" und:"und"`
	ElaDefOrUnd       option.Option[[]option.Option[string]] `json:",omitzero" und:"def,und"`
	ElaDefOrNull      option.Option[[]option.Option[string]] `json:",omitzero" und:"def,null"`
	ElaNullOrUnd      option.Option[conversion.Empty]        `json:",omitzero" und:"null,und"`
	ElaDefOrNullOrUnd elastic.Elastic[string]                `json:",omitzero" und:"def,null,und"`

	ElaEqEq option.Option[string]   `json:",omitzero" und:"len==1"`
	ElaGr   []option.Option[string] `json:",omitzero" und:"len>1"`
	ElaGrEq []option.Option[string] `json:",omitzero" und:"len>=1"`
	ElaLe   []option.Option[string] `json:",omitzero" und:"len<1"`
	ElaLeEq []option.Option[string] `json:",omitzero" und:"len<=1"`

	ElaEqEquRequired [2]option.Option[string]                `json:",omitzero" und:"required,len==2"`
	ElaEqEquNullish  und.Und[[2]option.Option[string]]       `json:",omitzero" und:"nullish,len==2"`
	ElaEqEquDef      [2]option.Option[string]                `json:",omitzero" und:"def,len==2"`
	ElaEqEquNull     option.Option[[2]option.Option[string]] `json:",omitzero" und:"null,len==2"`
	ElaEqEquUnd      option.Option[[2]option.Option[string]] `json:",omitzero" und:"und,len==2"`

	ElaEqEqNonNullSlice      und.Und[[]string]        `json:",omitzero" und:"values:nonnull"`
	ElaEqEqNonNullNullSlice  conversion.Empty         `json:",omitzero" und:"null,values:nonnull"`
	ElaEqEqNonNullSingle     string                   `json:",omitzero" und:"values:nonnull,len==1"`
	ElaEqEqNonNullNullSingle option.Option[string]    `json:",omitzero" und:"null,values:nonnull,len==1"`
	ElaEqEqNonNull           [3]string                `json:",omitzero" und:"values:nonnull,len==3"`
	ElaEqEqNonNullNull       option.Option[[3]string] `json:",omitzero" und:"null,values:nonnull,len==3"`
}

//undgen:generated
type WithTypeParamPlain[T any] struct {
	Foo string
	Bar T
	Baz T `json:",omitzero" und:"required"`
}
