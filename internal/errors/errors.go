package errors

import "errors"

var (
	// ErrInvalidRequestID is returned whenever the provided request ID is
	// invalid. For example, the request ID could have already been used for an
	// operation with a different change type.
	ErrInvalidRequestID = errors.New("invalid request ID")
	// ErrNotFound is returned whenever a resource cannot be found. For example,
	// a resource may not have been found within a database.
	ErrNotFound = errors.New("requested resource not found")
)
