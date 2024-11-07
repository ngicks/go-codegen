package targettypes

import (
	"fmt"

	"github.com/ngicks/und/undtag"
	"github.com/ngicks/und/validate"
)

//undgen:generated
func (v All) UndValidate() (err error) {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.OptRequired) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptRequired))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptNullish))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptDef))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptDefOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptDefOrNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptNullOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.OptDefOrNullOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndRequired))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndNullish))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndDef))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndDefOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndDefOrNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndNullOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.UndDefOrNullOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaRequired))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaNullish))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaDef))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaDefOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaDefOrNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaNullOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaDefOrNullOrUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEq))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaGr))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaGrEq))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaLe))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaLeEq))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquRequired))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquNullish))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquDef))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEquUnd))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullSlice))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullNullSlice))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullSingle))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullNullSingle))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
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
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.ElaEqEqNonNullNull))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"ElaEqEqNonNullNull",
			)
		}
	}
	return
}

//undgen:generated
func (v WithTypeParam[T]) UndValidate() (err error) {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.Baz) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Baz))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Baz",
			)
		}
	}
	return
}
