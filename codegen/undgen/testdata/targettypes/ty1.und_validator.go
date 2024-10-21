package targettypes

import (
	"fmt"

	"github.com/ngicks/und/undtag"
	"github.com/ngicks/und/validate"
)

//undgen:generated
func (v All) UndValidate() error {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptRequired) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptRequired",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptNullish) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptNullish",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptDef) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptDef",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptNull",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Und: true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
				Und: true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptDefOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptDefOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptDefOrNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptDefOrNull",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptNullOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptNullOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptDefOrNullOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"OptDefOrNullOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndRequired) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndRequired",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndNullish) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndNullish",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndDef) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndDef",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndNull",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Und: true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
				Und: true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndDefOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndDefOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndDefOrNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndDefOrNull",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndNullOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndNullOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidUnd(v.UndDefOrNullOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"UndDefOrNullOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaRequired) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaRequired",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaNullish) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaNullish",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaDef) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaDef",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaNull",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Und: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
				Und: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaDefOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaDefOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaDefOrNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaDefOrNull",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaNullOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaNullOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
				Und:  true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaDefOrNullOrUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaDefOrNullOrUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpEqEq,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEq) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEq",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpGr,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaGr) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaGr",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpGrEq,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaGrEq) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaGrEq",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpLe,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaLe) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaLe",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpLeEq,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaLeEq) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaLeEq",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEquRequired) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEquRequired",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
				Und:  true,
			},
			Len: &undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEquNullish) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEquNullish",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEquDef) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEquDef",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
			},
			Len: &undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEquNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEquNull",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
				Und: true,
			},
			Len: &undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEquUnd) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEquUnd",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEqNonNullSlice) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEqNonNullSlice",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEqNonNullNullSlice) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEqNonNullNullSlice",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpEqEq,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEqNonNullSingle) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEqNonNullSingle",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
			},
			Len: &undtag.LenValidator{
				Len: 1,
				Op:  undtag.LenOpEqEq,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEqNonNullNullSingle) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEqNonNullNullSingle",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 3,
				Op:  undtag.LenOpEqEq,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEqNonNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEqNonNull",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def:  true,
				Null: true,
			},
			Len: &undtag.LenValidator{
				Len: 3,
				Op:  undtag.LenOpEqEq,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.ElaEqEqNonNullNull) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"ElaEqEqNonNullNull",
			)
		}
	}

	return nil
}

//undgen:generated
func (v WithTypeParam[T]) UndValidate() error {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.Baz) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s", validator),
				"Baz",
			)
		}
	}

	return nil
}
