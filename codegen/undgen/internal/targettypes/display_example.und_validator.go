package targettypes

import (
	"fmt"

	"github.com/ngicks/und/undtag"
	"github.com/ngicks/und/validate"
)

//undgen:generated
func (v Example) UndValidate() (err error) {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.Bar) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Bar))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"bar",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidUnd(v.Baz) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Baz))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"baz",
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

		if !validator.ValidUnd(v.Qux) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Qux))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"qux",
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
		}.Into()

		if !validator.ValidElastic(v.Quux) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Quux))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"quux",
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
				Op:  undtag.LenOpGr,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.Corge) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Corge))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"corge",
			)
		}
	}
	return
}
