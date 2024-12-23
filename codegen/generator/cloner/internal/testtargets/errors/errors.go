package errors

// nolint
type includesError struct {
	err  error
	errs []error
}

// nolint
type errs []error
