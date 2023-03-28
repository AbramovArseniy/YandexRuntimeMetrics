package myerrors

import "errors"

var ErrTypeNotImplemented = errors.New("not implemented: ")
var ErrTypeBadRequest = errors.New("bad request: ")
var ErrTypeNotFound = errors.New("not found: ")
