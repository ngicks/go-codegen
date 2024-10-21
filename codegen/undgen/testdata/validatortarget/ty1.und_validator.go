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
				fmt.Errorf("%s", validator),
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
				fmt.Errorf("%s", validator),
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
				fmt.Errorf("%s", validator),
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
				fmt.Errorf("%s", validator),
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
				fmt.Errorf("%s", validator),
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
				return validate.AppendValidationErrorIndex(
					fmt.Errorf("%s", validator),
					fmt.Sprintf("%v", i),
				)
			}
			if err := option.UndValidate(val); err != nil {
				return validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", i),
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
				return validate.AppendValidationErrorIndex(
					fmt.Errorf("%s", validator),
					fmt.Sprintf("%v", i),
				)
			}
			if err := und.UndValidate(val); err != nil {
				return validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", i),
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
				return validate.AppendValidationErrorIndex(
					fmt.Errorf("%s", validator),
					fmt.Sprintf("%v", i),
				)
			}
			if err := elastic.UndValidate(val); err != nil {
				return validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", i),
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
				fmt.Errorf("%s", validator),
				"O",
			)
		}
		if err := option.UndValidate(v.O); err != nil {
			return validate.AppendValidationErrorIndex(
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
				return validate.AppendValidationErrorIndex(
					fmt.Errorf("%s", validator),
					fmt.Sprintf("%v", i),
				)
			}
			if err := option.UndValidate(val); err != nil {
				return validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", i),
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
				return validate.AppendValidationErrorIndex(
					fmt.Errorf("%s", validator),
					fmt.Sprintf("%v", i),
				)
			}
			if err := und.UndValidate(val); err != nil {
				return validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", i),
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
				return validate.AppendValidationErrorIndex(
					fmt.Errorf("%s", validator),
					fmt.Sprintf("%v", i),
				)
			}
			if err := elastic.UndValidate(val); err != nil {
				return validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", i),
				)
			}
		}
	}

	return nil
}
