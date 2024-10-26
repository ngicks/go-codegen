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
func (v All) UndValidate() error {
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
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
				Und: true,
			},
		}.Into()

		if !validator.ValidUnd(v.Qux) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Qux)),
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
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Quux)),
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
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Corge)),
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
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Grault)),
				"Grault",
			)
		}
	}

	return nil
}

//undgen:generated
func (v MapSliceArray) UndValidate() error {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		for i, val := range v.Foo {
			if !validator.ValidOpt(val) {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(i)),
						fmt.Sprintf("%v", i),
					),
					"foo",
				)
			}
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
			},
		}.Into()

		for i, val := range v.Bar {
			if !validator.ValidUnd(val) {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(i)),
						fmt.Sprintf("%v", i),
					),
					"bar",
				)
			}
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

		for i, val := range v.Baz {
			if !validator.ValidElastic(val) {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(i)),
						fmt.Sprintf("%v", i),
					),
					"baz",
				)
			}
		}
	}

	return nil
}

//undgen:generated
func (v ContainsImplementor) UndValidate() error {
	if err := v.I.UndValidate(); err != nil {
		return validate.AppendValidationErrorDot(err, "I")
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.O) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.O)),
				"O",
			)
		}
		if err := option.UndValidate(v.O); err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"O",
			)
		}
	}

	return nil
}

//undgen:generated
func (v MapSliceArrayContainsImplementor) UndValidate() error {
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		for i, val := range v.Foo {
			if !validator.ValidOpt(val) {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(i)),
						fmt.Sprintf("%v", i),
					),
					"Foo",
				)
			}
			if err := option.UndValidate(val); err != nil {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						err,
						fmt.Sprintf("%v", i),
					),
					"Foo",
				)
			}
		}
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Null: true,
			},
		}.Into()

		for i, val := range v.Bar {
			if !validator.ValidUnd(val) {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(i)),
						fmt.Sprintf("%v", i),
					),
					"Bar",
				)
			}
			if err := und.UndValidate(val); err != nil {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						err,
						fmt.Sprintf("%v", i),
					),
					"Bar",
				)
			}
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

		for i, val := range v.Baz {
			if !validator.ValidElastic(val) {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(i)),
						fmt.Sprintf("%v", i),
					),
					"Baz",
				)
			}
			if err := elastic.UndValidate(val); err != nil {
				return validate.AppendValidationErrorDot(
					validate.AppendValidationErrorIndex(
						err,
						fmt.Sprintf("%v", i),
					),
					"Baz",
				)
			}
		}
	}

	return nil
}
