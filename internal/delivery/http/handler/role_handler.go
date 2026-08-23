package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type RoleHandler struct {
	roles *usecase.RoleUsecase
}

func NewRoleHandler(roles *usecase.RoleUsecase) *RoleHandler {
	return &RoleHandler{roles: roles}
}

func (h *RoleHandler) List(c *gin.Context) {
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 10, 1, 100)

	roles, total, err := h.roles.List(c.Request.Context(), page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]gin.H, 0, len(roles))
	for _, r := range roles {
		data = append(data, gin.H{
			"id":   r.ID,
			"code": r.Name,
		})
	}

	response.SuccessWithMeta(c, http.StatusOK, "Roles retrieved", data, paginationMeta(page, limit, total))
}

type assignRolesRequest struct {
	UserID  string   `json:"user_id" binding:"required,uuid"`
	RoleIDs []string `json:"role_ids" binding:"required,gt=0,dive,uuid"`
}

func (h *RoleHandler) Assign(c *gin.Context) {
	var req assignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
		return
	}
	roleIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
	for _, raw := range req.RoleIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("Invalid role id: "+raw, err))
			return
		}
		roleIDs = append(roleIDs, id)
	}

	if err := h.roles.AssignToUser(c.Request.Context(), userID, roleIDs); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Role di-assign", nil)
}

func clampInt(raw string, fallback, min, max int) int {
	v, err := strconv.Atoi(raw)
	if err != nil || v < min {
		return fallback
	}
	if v > max {
		return max
	}
	return v
}
