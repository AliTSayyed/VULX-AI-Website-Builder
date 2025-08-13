/*
* This package will define custom errors in our applicaiton
* any error with the application must be tied (chained) to the domain
 */
package domain

import (
	"errors"
	"fmt"
)

type ErrorType int

const (
	ErrorTypeUnauthenticated ErrorType = iota
	ErrorTypePermissionDenied
	ErrorTypeInvalid
	ErrorTypeNotFound
	ErrorTypeAlreadyExists

	ErrorTypeInternal
	ErrorTypeUnimplemented
	ErrorTypeUnavailable
	ErrorTypeTimeout

	ErrorTypeUnknown
)

// custom error type that contians the domain error category and actual error message
type Error struct {
	t ErrorType
	e error
}

func NewError(t ErrorType, err error) *Error {
	return &Error{
		t: t,
		e: err,
	}
}

// if the error passed in is a custom Error Struct, then check to over write the t with the t in the error
func WrapError(msg string, err error) *Error {
	if err == nil {
		return nil
	}

	t := ErrorTypeUnknown
	var domainError *Error
	if errors.As(err, &domainError) {
		t = domainError.t
	}
	return &Error{t: t, e: fmt.Errorf(msg+": %w", err)}
}

func (e *Error) Type() ErrorType {
	return e.t
}

func (e *Error) Error() string {
	return e.e.Error()
}

func (e *Error) Unwrap() error {
	return e.e
}
