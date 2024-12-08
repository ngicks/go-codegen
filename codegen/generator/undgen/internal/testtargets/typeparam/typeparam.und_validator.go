// Code generated by github.com/ngicks/go-codegen/codegen DO NOT EDIT.
// to regenerate the code, refer to help by invoking
// go run github.com/ngicks/go-codegen/codegen undgen validator --help
package typeparam

import (
	"fmt"

	"github.com/ngicks/und/undtag"
	"github.com/ngicks/und/validate"
)

//codegen:generated
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
	{
		validator := undtag.UndOptExport{
			States: &undtag.StateValidator{
				Def: true,
			},
			Len: &undtag.LenValidator{
				Len: 2,
				Op:  undtag.LenOpEqEq,
			},
			Values: &undtag.ValuesValidator{
				Nonnull: true,
			},
		}.Into()

		v := v.Qux

		for k, v := range v {
			if !validator.ValidElastic(v) {
				err = fmt.Errorf("%s: value is %s", validator.Describe(), validate.ReportState(v))
			}
			if err != nil {
				err = validate.AppendValidationErrorIndex(
					err,
					fmt.Sprintf("%v", k),
				)
				break
			}
		}

		if err != nil {
			return validate.AppendValidationErrorDot(
				err,
				"qux",
			)
		}
	}
	return
}
