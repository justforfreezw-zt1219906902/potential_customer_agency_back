package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func BadRequest(message string, err error) *AppError {
	return &AppError{
		Status:  http.StatusBadRequest,
		Code:    "bad_request",
		Message: message,
		Err:     err,
	}
}

func ExternalService(message string, err error) *AppError {
	return &AppError{
		Status:  http.StatusBadGateway,
		Code:    "external_service_error",
		Message: message,
		Err:     err,
	}
}

func Internal(message string, err error) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: message,
		Err:     err,
	}
}

func NotFound(message string, err error) *AppError {
	return &AppError{Status: http.StatusNotFound, Code: "not_found", Message: message, Err: err}
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
