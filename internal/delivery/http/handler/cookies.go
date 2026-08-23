package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/asepzainudin14/mcbt/internal/config"
)

const RefreshTokenCookie = "refresh_token"

type CookieManager struct {
	cfg *config.Config
}

func NewCookieManager(cfg *config.Config) *CookieManager {
	return &CookieManager{cfg: cfg}
}

func (cm *CookieManager) sameSite() http.SameSite {
	switch cm.cfg.CookieSameSite {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}

func (cm *CookieManager) SetAuthCookies(c *gin.Context, accessToken, refreshToken, csrfToken string) {
	secure := cm.cfg.CookieSecure
	sameSite := cm.sameSite()

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(cm.cfg.JWTAccessTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     "/api/v1/auth",
		MaxAge:   int(cm.cfg.JWTRefreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		MaxAge:   int(cm.cfg.JWTRefreshTTL.Seconds()),
		HttpOnly: false,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func (cm *CookieManager) ClearAuthCookies(c *gin.Context) {
	sameSite := cm.sameSite()
	expired := time.Unix(0, 0)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  expired,
		HttpOnly: true,
		Secure:   cm.cfg.CookieSecure,
		SameSite: sameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		Expires:  expired,
		HttpOnly: true,
		Secure:   cm.cfg.CookieSecure,
		SameSite: sameSite,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  expired,
		HttpOnly: false,
		Secure:   cm.cfg.CookieSecure,
		SameSite: sameSite,
	})
}
