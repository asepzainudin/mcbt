package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
)

func ErrorHandler(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *apperror.AppError
		if !errors.As(err, &appErr) {
			appErr = mapUnknownError(err)
		}

		logAttrs := []any{
			slog.String("request_id", GetRequestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status_code", appErr.Code),
			slog.String("error", appErr.Error()),
		}
		if appErr.Err != nil {
			logAttrs = append(logAttrs, slog.String("cause", appErr.Err.Error()))
		}

		if appErr.Code >= http.StatusInternalServerError {
			log.Error("request_error", logAttrs...)
		} else {
			log.Warn("request_error", logAttrs...)
		}

		body := response.ErrorBody{
			Success: false,
			Message: appErr.Message,
		}
		if len(appErr.Details) > 0 {
			body.Errors = appErr.Details
		}

		c.AbortWithStatusJSON(appErr.Code, body)
	}
}

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic_recovered",
					slog.String("request_id", GetRequestID(c)),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorBody{
					Success: false,
					Message: "An unexpected error occurred",
				})
			}
		}()
		c.Next()
	}
}

func mapUnknownError(err error) *apperror.AppError {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validation failed",
			Err:     err,
			Details: validationDetails(validationErrs),
		}
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return apperror.BadRequest("Invalid request payload", err)
	}

	return apperror.Internal(err)
}

func validationDetails(errs validator.ValidationErrors) map[string]string {
	details := make(map[string]string, len(errs))
	for _, fe := range errs {
		field := toSnakeCase(fe.Field())
		switch fe.Tag() {
		case "required":
			details[field] = field + " wajib diisi"
		case "min":
			details[field] = field + " minimal " + fe.Param() + " karakter"
		case "max":
			details[field] = field + " maksimal " + fe.Param() + " karakter"
		case "uuid":
			details[field] = field + " harus berupa UUID yang valid"
		default:
			details[field] = field + " tidak valid (" + fe.Tag() + ")"
		}
	}
	return details
}

func toSnakeCase(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if !unicode.IsUpper(r) {
			b.WriteRune(r)
			continue
		}
		prevLower := i > 0 && unicode.IsLower(runes[i-1])
		prevUpper := i > 0 && unicode.IsUpper(runes[i-1])
		nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
		if i > 0 && (prevLower || (prevUpper && nextLower)) {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
