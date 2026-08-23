package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type SubjectHandler struct {
	uc *usecase.SubjectUsecase
}

func NewSubjectHandler(uc *usecase.SubjectUsecase) *SubjectHandler {
	return &SubjectHandler{uc: uc}
}

func (h *SubjectHandler) List(c *gin.Context) {
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 10, 1, 100)

	items, total, err := h.uc.List(c.Request.Context(), c.Query("search"), page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Mapel ditemukan", items, paginationMeta(page, limit, total))
}

type subjectRequest struct {
	Code        string  `json:"code" binding:"required,max=20"`
	Name        string  `json:"name" binding:"required,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

func (h *SubjectHandler) Create(c *gin.Context) {
	req, ok := bindSubjectRequest(c)
	if !ok {
		return
	}

	subject, err := h.uc.Create(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, "Mapel dibuat", subject)
}

func (h *SubjectHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	req, ok := bindSubjectRequest(c)
	if !ok {
		return
	}

	subject, err := h.uc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Mapel diperbarui", subject)
}

func (h *SubjectHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Mapel dihapus", nil)
}

func bindSubjectRequest(c *gin.Context) (usecase.SubjectInput, bool) {
	var req subjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return usecase.SubjectInput{}, false
	}
	return usecase.SubjectInput{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}, true
}
