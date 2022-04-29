package errors

import "errors"

var (
	ErrInvalidRequestID = errors.New("invalid request ID")
	ErrNotFound         = errors.New("requested resource not found")
)
