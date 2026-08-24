package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type ExamScheduleHandler struct {
	uc *usecase.ExamScheduleUsecase
}

func NewExamScheduleHandler(uc *usecase.ExamScheduleUsecase) *ExamScheduleHandler {
	return &ExamScheduleHandler{uc: uc}
}

func scheduleResponse(s *model.ExamSchedule) ginH {
	return ginH{
		"id":         s.ID,
		"exam_id":    s.ExamID,
		"start_time": s.StartTime,
		"end_time":   s.EndTime,
		"token":      s.Token,
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
	}
}

type examScheduleRequest struct {
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Token     string `json:"token"`
}

func bindScheduleRequest(c *gin.Context) (usecase.ExamScheduleInput, bool) {
	var req examScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return usecase.ExamScheduleInput{}, false
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.Error(&apperror.AppError{
			Code:    http.StatusUnprocessableEntity,
			Message: "Validasi gagal",
			Details: map[string]string{"start_time": "format harus RFC3339, contoh: 2026-09-01T08:00:00Z"},
		})
		return usecase.ExamScheduleInput{}, false
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.Error(&apperror.AppError{
			Code:    http.StatusUnprocessableEntity,
			Message: "Validasi gagal",
			Details: map[string]string{"end_time": "format harus RFC3339"},
		})
		return usecase.ExamScheduleInput{}, false
	}

	return usecase.ExamScheduleInput{StartTime: start, EndTime: end, Token: req.Token}, true
}

func (h *ExamScheduleHandler) Create(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	in, ok := bindScheduleRequest(c)
	if !ok {
		return
	}
	schedule, err := h.uc.Create(c.Request.Context(), examID, in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Jadwal ujian dibuat", scheduleResponse(schedule))
}

// GetByExam returns the single schedule of an exam (data null bila belum ada).
func (h *ExamScheduleHandler) GetByExam(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	schedule, err := h.uc.GetByExam(c.Request.Context(), examID)
	if err != nil {
		c.Error(err)
		return
	}
	if schedule == nil {
		response.Success(c, http.StatusOK, "Belum ada jadwal", nil)
		return
	}
	response.Success(c, http.StatusOK, "Jadwal ujian", scheduleResponse(schedule))
}

func (h *ExamScheduleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	in, ok := bindScheduleRequest(c)
	if !ok {
		return
	}
	schedule, err := h.uc.Update(c.Request.Context(), id, in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Jadwal diperbarui", scheduleResponse(schedule))
}

func (h *ExamScheduleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Jadwal dihapus", nil)
}

// GenerateToken regenerates the exam token of a schedule.
func (h *ExamScheduleHandler) GenerateToken(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	schedule, token, err := h.uc.GenerateToken(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Token baru berhasil dibuat", gin.H{
		"token":       token,
		"schedule_id": schedule.ID,
	})
}
