package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type ClassHandler struct {
	uc *usecase.ClassUsecase
}

func NewClassHandler(uc *usecase.ClassUsecase) *ClassHandler {
	return &ClassHandler{uc: uc}
}

func (h *ClassHandler) List(c *gin.Context) {
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 10, 1, 100)

	var ayID *uuid.UUID
	if raw := c.Query("academic_year_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("academic_year_id harus UUID yang valid", err))
			return
		}
		ayID = &parsed
	}

	items, total, err := h.uc.List(c.Request.Context(), c.Query("search"), ayID, page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, "Kelas ditemukan", items, paginationMeta(page, limit, total))
}

type classRequest struct {
	Name           string `json:"name" binding:"required,max=100"`
	AcademicYearID string `json:"academic_year_id" binding:"required,uuid"`
}

func (h *ClassHandler) Create(c *gin.Context) {
	in, err := bindClassRequest(c)
	if err != nil {
		return
	}

	class, err := h.uc.Create(c.Request.Context(), in)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusCreated, "Kelas dibuat", class)
}

func (h *ClassHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	in, err := bindClassRequest(c)
	if err != nil {
		return
	}

	class, err := h.uc.Update(c.Request.Context(), id, in)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Kelas diperbarui", class)
}

func (h *ClassHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, http.StatusOK, "Kelas dihapus", nil)
}

func bindClassRequest(c *gin.Context) (usecase.ClassInput, error) {
	var req classRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return usecase.ClassInput{}, err
	}
	ayID, _ := uuid.Parse(req.AcademicYearID)
	return usecase.ClassInput{Name: req.Name, AcademicYearID: ayID}, nil
}
