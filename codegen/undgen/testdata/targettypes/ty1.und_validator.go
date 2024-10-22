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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.opt_required)),
				"opt_required",
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptNullish)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptDef)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptDefOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptDefOrNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptNullOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptDefOrNullOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndRequired)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndNullish)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndDef)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndDefOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndDefOrNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndNullOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndDefOrNullOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaRequired)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaNullish)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaDef)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaDefOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaDefOrNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaNullOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaDefOrNullOrUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEq)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaGr)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaGrEq)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaLe)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaLeEq)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquRequired)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquNullish)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquDef)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquUnd)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullSlice)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullNullSlice)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullSingle)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullNullSingle)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullNull)),
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
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Baz)),
				"Baz",
			)
		}
	}

	return nil
}
