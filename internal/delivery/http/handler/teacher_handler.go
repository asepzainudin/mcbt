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

type TeacherHandler struct {
	uc *usecase.TeacherUsecase
}

func NewTeacherHandler(uc *usecase.TeacherUsecase) *TeacherHandler {
	return &TeacherHandler{uc: uc}
}

func (h *TeacherHandler) List(c *gin.Context) {
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 20, 1, 100)

	items, total, err := h.uc.List(c.Request.Context(), c.Query("search"), page, limit)
	if err != nil {
		c.Error(err)
		return
	}
	response.SuccessWithMeta(c, http.StatusOK, "Guru ditemukan", items, paginationMeta(page, limit, total))
}

func (h *TeacherHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	teacher, err := h.uc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Detail guru", teacher)
}

type teacherRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=50"`
	Name     string  `json:"name" binding:"required,max=100"`
	Email    string  `json:"email" binding:"required,email"`
	Nip      *string `json:"nip" binding:"omitempty,max=30"`
	Phone    *string `json:"phone" binding:"omitempty,max=20"`
	Address  *string `json:"address" binding:"omitempty,max=255"`
}

func (h *TeacherHandler) Create(c *gin.Context) {
	var req teacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	teacher, err := h.uc.Create(c.Request.Context(), repository.TeacherUpsert{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Nip:      req.Nip,
		Phone:    req.Phone,
		Address:  req.Address,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Guru dibuat", teacher)
}

type teacherUpdateRequest struct {
	Name    string  `json:"name" binding:"required,max=100"`
	Email   string  `json:"email" binding:"required,email"`
	Nip     *string `json:"nip" binding:"omitempty,max=30"`
	Phone   *string `json:"phone" binding:"omitempty,max=20"`
	Address *string `json:"address" binding:"omitempty,max=255"`
}

func (h *TeacherHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	var req teacherUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	teacher, err := h.uc.Update(c.Request.Context(), id, repository.TeacherUpdate{
		Name:    req.Name,
		Email:   req.Email,
		Nip:     req.Nip,
		Phone:   req.Phone,
		Address: req.Address,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Guru diperbarui", teacher)
}

func (h *TeacherHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Guru dihapus", nil)
}

func (h *TeacherHandler) Import(c *gin.Context) {
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

func (h *TeacherHandler) Template(c *gin.Context) {
	data, err := h.uc.TemplateXLSX()
	if err != nil {
		c.Error(apperror.Internal(err))
		return
	}
	c.Header("Content-Disposition", "attachment; filename=data_guru_template.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ResetPassword: admin mengganti password guru (kosong = generate acak).
func (h *TeacherHandler) ResetPassword(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}
	_ = c.ShouldBindJSON(&req)

	result, err := h.uc.ResetPassword(c.Request.Context(), id, req.NewPassword)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Password guru direset", result)
}
