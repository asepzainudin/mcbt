package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type AcademicYearHandler struct {
	uc *usecase.AcademicYearUsecase
}

func NewAcademicYearHandler(uc *usecase.AcademicYearUsecase) *AcademicYearHandler {
	return &AcademicYearHandler{uc: uc}
}

func (h *AcademicYearHandler) List(c *gin.Context) {
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 10, 1, 100)

	items, total, err := h.uc.List(c.Request.Context(), c.Query("search"), page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Tahun ajaran ditemukan", items, paginationMeta(page, limit, total))
}

type academicYearRequest struct {
	Year     string `json:"year" binding:"required"`
	Semester string `json:"semester" binding:"required,oneof=ODD EVEN"`
}

func (h *AcademicYearHandler) Create(c *gin.Context) {
	var req academicYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	ay, err := h.uc.Create(c.Request.Context(), usecase.AcademicYearInput{
		Year:     req.Year,
		Semester: req.Semester,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, "Tahun ajaran dibuat", ay)
}

func (h *AcademicYearHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	var req academicYearRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	ay, err := h.uc.Update(c.Request.Context(), id, usecase.AcademicYearInput{
		Year:     req.Year,
		Semester: req.Semester,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Tahun ajaran diperbarui", ay)
}

func (h *AcademicYearHandler) Activate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	ay, err := h.uc.Activate(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Tahun ajaran diaktifkan", ay)
}

func (h *AcademicYearHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Tahun ajaran dihapus", nil)
}

func paginationMeta(page, limit int, total int64) response.Meta {
	totalPages := int64(0)
	if limit > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}
	return response.Meta{
		"page":        page,
		"limit":       limit,
		"total_items": total,
		"total_pages": totalPages,
	}
}
