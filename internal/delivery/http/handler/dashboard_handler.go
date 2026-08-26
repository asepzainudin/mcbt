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

type DashboardHandler struct {
	uc *usecase.DashboardUsecase
}

func NewDashboardHandler(uc *usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

// principalActor mengembalikan userID + status admin dari sesi aktif.
func principalActor(c *gin.Context) (uuid.UUID, bool, bool) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		return uuid.Nil, false, false
	}
	id, err := uuid.Parse(principal.UserID)
	if err != nil {
		return uuid.Nil, false, false
	}
	isAdmin := false
	for _, r := range principal.Roles {
		if r == "admin" {
			isAdmin = true
			break
		}
	}
	return id, isAdmin, true
}

func principalUserID(c *gin.Context) (uuid.UUID, bool) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(principal.UserID)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// Admin: GET /dashboard/admin
func (h *DashboardHandler) Admin(c *gin.Context) {
	stats, err := h.uc.Admin(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Ringkasan dashboard admin", stats)
}

// Teacher: GET /dashboard/teacher
func (h *DashboardHandler) Teacher(c *gin.Context) {
	userID, ok := principalUserID(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}
	stats, err := h.uc.Teacher(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Ringkasan dashboard guru", stats)
}

// Student: GET /dashboard/student
func (h *DashboardHandler) Student(c *gin.Context) {
	userID, ok := principalUserID(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}
	summary, err := h.uc.Student(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Ringkasan dashboard siswa", summary)
}
