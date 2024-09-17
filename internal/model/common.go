package model

type ValidationError struct {
	Inner error
}

func (err *ValidationError) Error() string {
	return err.Inner.Error()
}

func NewValidationError(err error) error {
	return &ValidationError{Inner: err}
}
