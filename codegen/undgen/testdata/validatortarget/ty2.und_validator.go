package validatortarget

import (
	"fmt"

	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/option"
	sliceelastic "github.com/ngicks/und/sliceund/elastic"
	"github.com/ngicks/und/undtag"
	"github.com/ngicks/und/validate"
)

//undgen:generated
func (v A) UndValidate() error {
	for i, val := range v {
		if err := elastic.UndValidate(val); err != nil {
			return validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", i),
			)
		}
	}

	return nil
}

//undgen:generated
func (v B) UndValidate() error {
	for i, val := range v {
		if err := sliceelastic.UndValidate(val); err != nil {
			return validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", i),
			)
		}
	}

	return nil
}

//undgen:generated
func (v C) UndValidate() error {
	for i, val := range v {
		if err := option.UndValidate(val); err != nil {
			return validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", i),
			)
		}
	}

	return nil
}

//undgen:generated
func (v D) UndValidate() error {
	if err := v.Foo.UndValidate(); err != nil {
		return validate.AppendValidationErrorDot(err, "Foo")
	}
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.Bar) {
			return validate.AppendValidationErrorDot(
				fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Bar)),
				"Bar",
			)
		}
		if err := option.UndValidate(v.Bar); err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Bar",
			)
		}
	}

	return nil
}
