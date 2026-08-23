package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	mw "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type AuthHandler struct {
	auth    *usecase.AuthUsecase
	cookies *CookieManager
}

func NewAuthHandler(auth *usecase.AuthUsecase, cookies *CookieManager) *AuthHandler {
	return &AuthHandler{auth: auth, cookies: cookies}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.auth.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	csrf, err := newCSRFToken()
	if err != nil {
		c.Error(err)
		return
	}

	h.cookies.SetAuthCookies(c, result.Token.AccessToken, result.Token.RefreshToken, csrf)

	response.Success(c, http.StatusOK, "Login berhasil", gin.H{
		"user_id":  result.User.ID,
		"username": result.User.Username,
		"roles":    roleNames(result.User.Roles),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.cookies.ClearAuthCookies(c)
	response.Success(c, http.StatusOK, "Logout berhasil", nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(RefreshTokenCookie)
	if err != nil {
		c.Error(apperror.New(http.StatusUnauthorized, "Refresh token missing", nil))
		return
	}

	result, err := h.auth.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.cookies.ClearAuthCookies(c)
		c.Error(err)
		return
	}

	csrf, err := newCSRFToken()
	if err != nil {
		c.Error(err)
		return
	}

	h.cookies.SetAuthCookies(c, result.Token.AccessToken, result.Token.RefreshToken, csrf)

	response.Success(c, http.StatusOK, "Token diperbarui", gin.H{
		"user_id":  result.User.ID,
		"username": result.User.Username,
		"roles":    roleNames(result.User.Roles),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}

	user, err := h.auth.Me(c.Request.Context(), parseUUID(principal.UserID))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Current user", gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"username": user.Username,
		"roles":    roleNames(user.Roles),
	})
}

type changePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		c.Error(apperror.New(422, "confirm_password tidak sama dengan new_password", nil))
		return
	}

	userID, err := uuid.Parse(principal.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
		return
	}

	if err := h.auth.ChangePassword(
		c.Request.Context(), userID, req.OldPassword, req.NewPassword,
	); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Password diperbarui", nil)
}
