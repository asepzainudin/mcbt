package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	CodeBadRequest    = http.StatusBadRequest
	CodeUnauthorized  = http.StatusUnauthorized
	CodeForbidden     = http.StatusForbidden
	CodeNotFound      = http.StatusNotFound
	CodeUnprocessable = http.StatusUnprocessableEntity
	CodeInternal      = http.StatusInternalServerError
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func BadRequest(message string, err error) *AppError {
	return &AppError{Code: CodeBadRequest, Message: defaultMsg(message, "Bad request"), Err: err}
}

func Unauthorized(message string, err error) *AppError {
	return &AppError{Code: CodeUnauthorized, Message: defaultMsg(message, "Unauthorized"), Err: err}
}

func Forbidden(message string, err error) *AppError {
	return &AppError{Code: CodeForbidden, Message: defaultMsg(message, "Forbidden"), Err: err}
}

func NotFound(message string, err error) *AppError {
	return &AppError{Code: CodeNotFound, Message: defaultMsg(message, "Resource not found"), Err: err}
}

func Unprocessable(message string, err error) *AppError {
	return &AppError{Code: CodeUnprocessable, Message: defaultMsg(message, "Validation failed"), Err: err}
}

func Internal(err error) *AppError {
	return &AppError{Code: CodeInternal, Message: "An unexpected error occurred", Err: err}
}

func From(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}

func defaultMsg(msg, fallback string) string {
	if msg == "" {
		return fallback
	}
	return msg
}
