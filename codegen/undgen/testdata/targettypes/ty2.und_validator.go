package targettypes

import (
	"github.com/ngicks/und/validate"
)

//undgen:generated
func (v IncludesSubTarget) UndValidate() error {
	if err := v.Foo.UndValidate(); err != nil {
		return validate.AppendValidationErrorDot(err, "Foo")
	}

	return nil
}
