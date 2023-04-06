// Module myerrors has types of errors
package myerrors

import "errors"

// ErrTypeNotImplemented for http status not implemented
var ErrTypeNotImplemented = errors.New("not implemented: ")

// ErrTypeNotImplemented for http status bad request
var ErrTypeBadRequest = errors.New("bad request: ")

// ErrTypeNotImplemented for http status not found
var ErrTypeNotFound = errors.New("not found: ")
