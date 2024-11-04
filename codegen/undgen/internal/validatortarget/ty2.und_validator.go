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
func (v A) UndValidate() (err error) {
LOOP:
	for k, v := range v {
		err = elastic.UndValidate(v)
		if err != nil {
			err = validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", k),
			)
			break LOOP
		}
	}
	return
}

//undgen:generated
func (v B) UndValidate() (err error) {
LOOP:
	for k, v := range v {
		err = sliceelastic.UndValidate(v)
		if err != nil {
			err = validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", k),
			)
			break LOOP
		}
	}
	return
}

//undgen:generated
func (v C) UndValidate() (err error) {
LOOP:
	for k, v := range v {
		err = option.UndValidate(v)
		if err != nil {
			err = validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", k),
			)
			break LOOP
		}
	}
	return
}

//undgen:generated
func (v D) UndValidate() (err error) {
	{
		err = v.Foo.UndValidate()
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
				Def: true,
			},
		}.Into()

		if !validator.ValidOpt(v.Bar) {
			err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v.Bar))
		}
		if err == nil {
			err = option.UndValidate(v.Bar)
		}

		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"Bar",
			)
		}
	}
	return
}
