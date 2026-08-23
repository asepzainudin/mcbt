package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"

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

		c.AbortWithStatusJSON(appErr.Code, response.ErrorBody{
			Success: false,
			Message: appErr.Message,
		})
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
		return apperror.Unprocessable("Validation failed", err)
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return apperror.BadRequest("Invalid request payload", err)
	}

	return apperror.Internal(err)
}
