package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type QuestionImportHandler struct {
	uc     *usecase.QuestionImportUsecase
	access *usecase.AccessUsecase
}

func NewQuestionImportHandler(uc *usecase.QuestionImportUsecase, access *usecase.AccessUsecase) *QuestionImportHandler {
	return &QuestionImportHandler{uc: uc, access: access}
}

func (h *QuestionImportHandler) Template(c *gin.Context) {
	data, err := h.uc.TemplateXLSX()
	if err != nil {
		c.Error(apperror.Internal(err))
		return
	}
	c.Header("Content-Disposition", "attachment; filename=soal_template.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// Validate accepts multipart fields: file=soal.xlsx & question_bank_id=<uuid>.
// Returns import_token on success, or 422 with per-row error report.
func (h *QuestionImportHandler) Validate(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error(apperror.BadRequest("File wajib diunggah pada field 'file'", err))
		return
	}

	bankID, err := uuid.Parse(c.PostForm("question_bank_id"))
	if err != nil {
		c.Error(apperror.New(http.StatusUnprocessableEntity,
			"question_bank_id wajib UUID yang valid", err))
		return
	}

	// guru hanya boleh impor ke bank miliknya
	if uid, isAdmin, ok := principalActor(c); ok && !isAdmin {
		if err := h.access.AssertBankOwner(c.Request.Context(), uid, false, bankID); err != nil {
			c.Error(err)
			return
		}
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

	token, totalRows, skipped, err := h.uc.Validate(c.Request.Context(), buf, bankID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"message": err.Error(),
			"data": gin.H{
				"total_rows": totalRows,
				"skipped":    skipped,
			},
		})
		return
	}
	response.SuccessWithMeta(c, http.StatusOK, "Validasi berhasil", gin.H{
		"import_token": token,
	}, response.Meta{
		"total_rows": totalRows,
	})
}

type processImportRequest struct {
	ImportToken string `json:"import_token" binding:"required"`
}

func (h *QuestionImportHandler) Process(c *gin.Context) {
	var req processImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.uc.Process(c.Request.Context(), req.ImportToken)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Import soal selesai", result)
}
