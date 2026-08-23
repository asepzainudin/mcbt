package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

const CSRFTokenCookie = "csrf_token"
const CSRFHeaderName = "X-CSRF-Token"

func CSRFProtection(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}
		if c.FullPath() == "/api/v1/auth/login" || c.Request.URL.Path == "/api/v1/auth/login" {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(CSRFTokenCookie)
		if err != nil || cookieToken == "" {
			log.Warn("csrf_rejected",
				slog.String("reason", "missing csrf cookie"),
				slog.String("request_id", GetRequestID(c)),
				slog.String("path", c.Request.URL.Path),
			)
			c.Error(apperror.New(http.StatusForbidden, "CSRF token missing", nil))
			c.Abort()
			return
		}

		headerToken := c.GetHeader(CSRFHeaderName)
		if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
			log.Warn("csrf_rejected",
				slog.String("reason", "csrf token mismatch"),
				slog.String("request_id", GetRequestID(c)),
				slog.String("path", c.Request.URL.Path),
			)
			c.Error(apperror.New(http.StatusForbidden, "CSRF token mismatch", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
