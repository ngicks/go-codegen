package validatortarget

import (
	"fmt"

	"github.com/ngicks/und"
	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
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
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
				Und: true,
			},
		}.Into()

		if !validator.ValidUnd(v.Qux) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Qux))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Qux",
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
		}.Into()

		if !validator.ValidElastic(v.Quux) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Quux))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Quux",
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

		if !validator.ValidUnd(v.Corge) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Corge))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Corge",
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
				Op:  undtag.LenOpGrEq,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		if !validator.ValidElastic(v.Grault) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Grault))
		}
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Grault",
			)
		}
	}
	return
}

//undgen:generated
func (v MapSliceArray) UndValidate() (err error) {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		v := v.Foo

	LOOP_Foo:
		for k, v := range v {
			if !validator.ValidOpt(v) {
				err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v))
			}
			if err != nil {
				err = validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", k),
				)
				break LOOP_Foo
			}
		}

		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"foo",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
			},
		}.Into()

		v := v.Bar

	LOOP_Bar:
		for k, v := range v {
			if !validator.ValidUnd(v) {
				err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v))
			}
			if err != nil {
				err = validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", k),
				)
				break LOOP_Bar
			}
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
				Und: true,
			},
			Len: &undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpGrEq,
			},
		}.Into()

		v := v.Baz

	LOOP_Baz:
		for k, v := range v {
			if !validator.ValidElastic(v) {
				err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v))
			}
			if err != nil {
				err = validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", k),
				)
				break LOOP_Baz
			}
		}

		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"baz",
			)
		}
	}
	return
}

//undgen:generated
func (v ContainsImplementor) UndValidate() (err error) {
	{
		err = v.I.UndValidate()
		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"I",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.O) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.O))
		}
		if err == nil {
			err = option.UndValidate(v.O)
		}

		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"O",
			)
		}
	}
	return
}

//undgen:generated
func (v MapSliceArrayContainsImplementor) UndValidate() (err error) {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		v := v.Foo

	LOOP_Foo:
		for k, v := range v {
			if !validator.ValidOpt(v) {
				err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v))
			}
			if err == nil {
				err = option.UndValidate(v)
			}

			if err != nil {
				err = validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", k),
				)
				break LOOP_Foo
			}
		}

		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Foo",
			)
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
			},
		}.Into()

		v := v.Bar

	LOOP_Bar:
		for k, v := range v {
			if !validator.ValidUnd(v) {
				err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v))
			}
			if err == nil {
				err = und.UndValidate(v)
			}

			if err != nil {
				err = validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", k),
				)
				break LOOP_Bar
			}
		}

		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Bar",
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
				Op:  undtag.LenOpGrEq,
			},
		}.Into()

		v := v.Baz

	LOOP_Baz:
		for k, v := range v {
			if !validator.ValidElastic(v) {
				err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v))
			}
			if err == nil {
				err = elastic.UndValidate(v)
			}

			if err != nil {
				err = validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", k),
				)
				break LOOP_Baz
			}
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
