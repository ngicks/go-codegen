package targettypes

import (
	"fmt"

	"github.com/ngicks/und/elastic"
	"github.com/ngicks/und/sliceund"
	sliceElastic "github.com/ngicks/und/sliceund/elastic"
	"github.com/ngicks/und/validate"
)

//undgen:generated
func (v A) UndValidate() error {

	return nil
}

//undgen:generated
func (v B) UndValidate() error {

	return nil
}

//undgen:generated
func (v C) UndValidate() error {
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
func (v D) UndValidate() error {
	for i, val := range v {
		if err := sliceund.UndValidate(val); err != nil {
			return validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", i),
			)
		}
	}

	return nil
}

//undgen:generated
func (v F) UndValidate() error {
	for i, val := range v {
		if err := sliceElastic.UndValidate(val); err != nil {
			return validate.AppendValidationErrorIndex(
				err,
				fmt.Sprintf("%v", i),
			)
		}
	}

	return nil
}

//undgen:generated
func (v Parametrized[T]) UndValidate() error {

	return nil
}

//undgen:generated
func (v IncludesSubTarget) UndValidate() error {
	if err := v.Foo.UndValidate(); err != nil {
		return validate.AppendValidationErrorDot(err, "Foo")
	}

	return nil
}

//undgen:generated
func (v NestedImplementor) UndValidate() error {

	return nil
}
