package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type ExportHandler struct {
	uc *usecase.ExportUsecase
}

func NewExportHandler(uc *usecase.ExportUsecase) *ExportHandler {
	return &ExportHandler{uc: uc}
}

func (h *ExportHandler) send(c *gin.Context, f *usecase.ExportFile) {
	c.Header("Content-Disposition", `attachment; filename="`+f.Filename+`"`)
	c.Data(http.StatusOK, f.ContentType, f.Data)
}

// ExamResults: GET /exams/:id/export?format=xlsx|pdf
func (h *ExportHandler) ExamResults(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	file, err := h.uc.ExamResults(c.Request.Context(), examID, c.Query("format"))
	if err != nil {
		c.Error(err)
		return
	}
	h.send(c, file)
}

// Students: GET /export/students?format=xlsx|pdf
func (h *ExportHandler) Students(c *gin.Context) {
	file, err := h.uc.Students(c.Request.Context(), c.Query("format"))
	if err != nil {
		c.Error(err)
		return
	}
	h.send(c, file)
}

// Teachers: GET /export/teachers?format=xlsx|pdf
func (h *ExportHandler) Teachers(c *gin.Context) {
	file, err := h.uc.Teachers(c.Request.Context(), c.Query("format"))
	if err != nil {
		c.Error(err)
		return
	}
	h.send(c, file)
}

var _ = time.Now
