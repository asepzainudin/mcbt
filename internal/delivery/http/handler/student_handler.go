package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/repository"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type StudentHandler struct {
	uc *usecase.StudentUsecase
}

func NewStudentHandler(uc *usecase.StudentUsecase) *StudentHandler {
	return &StudentHandler{uc: uc}
}

func (h *StudentHandler) List(c *gin.Context) {
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 20, 1, 100)

	var classID *uuid.UUID
	if raw := c.Query("class_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("class_id harus UUID yang valid", err))
			return
		}
		classID = &parsed
	}

	items, total, err := h.uc.List(c.Request.Context(), c.Query("search"), classID, page, limit)
	if err != nil {
		c.Error(err)
		return
	}
	response.SuccessWithMeta(c, http.StatusOK, "Siswa ditemukan", items, paginationMeta(page, limit, total))
}

func (h *StudentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	student, err := h.uc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Detail siswa", student)
}

type studentRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=50"`
	Name     string  `json:"name" binding:"required,max=100"`
	Email    string  `json:"email" binding:"required,email"`
	Nis      string  `json:"nis" binding:"required,max=30"`
	ClassID  *string `json:"class_id" binding:"omitempty,uuid"`
	Phone    *string `json:"phone" binding:"omitempty,max=20"`
	Address  *string `json:"address" binding:"omitempty,max=255"`
}

func parseClassID(raw *string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, apperror.BadRequest("class_id tidak valid", err)
	}
	return &id, nil
}

func (h *StudentHandler) Create(c *gin.Context) {
	var req studentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	classID, err := parseClassID(req.ClassID)
	if err != nil {
		c.Error(err)
		return
	}

	student, err := h.uc.Create(c.Request.Context(), repository.StudentUpsert{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Nis:      req.Nis,
		ClassID:  classID,
		Phone:    req.Phone,
		Address:  req.Address,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Siswa dibuat", student)
}

type studentUpdateRequest struct {
	Name    string  `json:"name" binding:"required,max=100"`
	Email   string  `json:"email" binding:"required,email"`
	Nis     string  `json:"nis" binding:"required,max=30"`
	ClassID *string `json:"class_id" binding:"omitempty,uuid"`
	Phone   *string `json:"phone" binding:"omitempty,max=20"`
	Address *string `json:"address" binding:"omitempty,max=255"`
}

func (h *StudentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	var req studentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	classID, err := parseClassID(req.ClassID)
	if err != nil {
		c.Error(err)
		return
	}

	student, err := h.uc.Update(c.Request.Context(), id, repository.StudentUpdate{
		Name:    req.Name,
		Email:   req.Email,
		Nis:     req.Nis,
		ClassID: classID,
		Phone:   req.Phone,
		Address: req.Address,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Siswa diperbarui", student)
}

func (h *StudentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Siswa dihapus", nil)
}

type changeClassRequest struct {
	TargetClassID string `json:"target_class_id" binding:"required,uuid"`
}

func (h *StudentHandler) ChangeClass(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	var req changeClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	targetID, _ := uuid.Parse(req.TargetClassID)
	student, err := h.uc.ChangeClass(c.Request.Context(), id, targetID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Kelas siswa diperbarui", student)
}

func (h *StudentHandler) ResetPassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}
	_ = c.ShouldBindJSON(&req) // opsional: kosong = generate acak

	result, err := h.uc.ResetPassword(c.Request.Context(), id, req.NewPassword)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Password siswa direset", result)
}

func (h *StudentHandler) Import(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error(apperror.BadRequest("File wajib diunggah pada field 'file'", err))
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.Error(apperror.BadRequest("Gagal membaca file", err))
		return
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		c.Error(apperror.BadRequest("Gagal membaca isi file", err))
		return
	}

	result, err := h.uc.Import(c.Request.Context(), buf)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Import selesai", result)
}

func (h *StudentHandler) Template(c *gin.Context) {
	data, err := h.uc.TemplateXLSX()
	if err != nil {
		c.Error(apperror.Internal(err))
		return
	}
	c.Header("Content-Disposition", "attachment; filename=data_siswa_template.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
