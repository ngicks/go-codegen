package targettypes

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

	return nil
}

//undgen:generated
func (v D) UndValidate() error {

	return nil
}

//undgen:generated
func (v F) UndValidate() error {

	return nil
}

//undgen:generated
func (v Parametrized[T]) UndValidate() error {

	return nil
}

//undgen:generated
func (v IncludesSubTarget) UndValidate() error {
	if err := v.Foo.UndValidate(); err != nil {
		return err
	}

	return nil
}

//undgen:generated
func (v NestedImplementor) UndValidate() error {

	return nil
}
