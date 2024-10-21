package validatortarget

import (
	"fmt"

	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/validate"
)

//undgen:generated
func (v A) UndValidate() error {
	for i, val := range v {
		if err := val.UndValidate(); err != nil {
			return validate.AppendValidationErrorIndex(
				err, fmt.Sprintf("%v", i),
			)
		}
		if err := elastic.UndValidate(val); err != nil {
			return validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", i),
			)
		}
	}

	return nil
}
