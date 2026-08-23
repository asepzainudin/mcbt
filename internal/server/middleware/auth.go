package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	jwtmanager "github.com/asepzainudin14/mcbt/internal/pkg/jwt"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

const (
	AccessTokenCookie = "access_token"
	PrincipalKey      = "auth_principal"
)

type Principal struct {
	UserID       string
	Username     string
	Name         string
	Roles        []string
	TokenVersion int
}

func (p *Principal) HasRole(role string) bool {
	for _, r := range p.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func CurrentPrincipal(c *gin.Context) (*Principal, bool) {
	v, exists := c.Get(PrincipalKey)
	if !exists {
		return nil, false
	}
	p, ok := v.(*Principal)
	return p, ok
}

func Authenticate(log *slog.Logger, tokens *jwtmanager.Manager, users *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie(AccessTokenCookie)
		if err != nil {
			c.Error(apperror.New(401, "Authentication required", nil))
			c.Abort()
			return
		}

		claims, err := tokens.Parse(tokenStr, jwtmanager.TokenTypeAccess)
		if err != nil {
			c.Error(apperror.New(401, "Invalid or expired session", err))
			c.Abort()
			return
		}

		user, err := users.FindByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.Error(apperror.Internal(err))
			c.Abort()
			return
		}
		if user == nil {
			c.Error(apperror.New(401, "Invalid session", nil))
			c.Abort()
			return
		}
		if !user.IsActive {
			c.Error(apperror.New(403, "Account is disabled", nil))
			c.Abort()
			return
		}
		if user.TokenVersion != claims.TokenVersion {
			c.Error(apperror.New(401, "Session revoked, please login again", nil))
			c.Abort()
			return
		}

		principal := &Principal{
			UserID:       user.ID.String(),
			Username:     user.Username,
			Name:         user.Name,
			TokenVersion: user.TokenVersion,
		}
		for _, role := range user.Roles {
			principal.Roles = append(principal.Roles, role.Name)
		}

		c.Set(PrincipalKey, principal)
		c.Next()
	}
}

func RequireRoles(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := CurrentPrincipal(c)
		if !ok {
			c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
			c.Abort()
			return
		}

		for _, role := range allowed {
			if principal.HasRole(role) {
				c.Next()
				return
			}
		}

		c.Error(apperror.New(http.StatusForbidden, "Insufficient permissions", nil))
		c.Abort()
	}
}
