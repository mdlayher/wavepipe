package models

import (
	"fmt"
)

// Validator provides the Validate method, which ensures that fields on a struct
// contain valid values.  An error is returned if any values are not valid.
type Validator interface {
	Validate() error
}

// EmptyFieldError is returned when a field fails a call to Validate
// due to empty input data.
//
// The struct contains the name of the field which failed.
type EmptyFieldError struct {
	Field string
}

// Error returns a string representation of an EmptyFieldError.
func (e *EmptyFieldError) Error() string {
	return fmt.Sprintf("empty field: %s", e.Field)
}

// InvalidFieldError is returned when a field fails a call to Validate
// due to invalid input data.
//
// The struct contains the name of the field which failed, human-readable
// details regarding its failure, and if possible, the error which caused
// the failure to be triggered.
type InvalidFieldError struct {
	Field   string
	Err     error
	Details string
}

// Error returns a string representation of an InvalidFieldError.
func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("invalid field: %s (%s)", e.Field, e.Details)
}
