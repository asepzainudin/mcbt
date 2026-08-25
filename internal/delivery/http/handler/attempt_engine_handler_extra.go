package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

// Heartbeat: sinkronisasi sisa waktu berdasarkan jam server.
func (h *AttemptEngineHandler) Heartbeat(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}

	attempt, hb, err := h.uc.Heartbeat(c.Request.Context(), userID, attemptID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Heartbeat", ginH{
		"server_time":       hb.ServerTime,
		"remaining_seconds": hb.RemainingSeconds,
		"is_expired":        hb.IsExpired,
		"attempt_status":    attempt.Status,
		"submitted_at":      attempt.SubmittedAt,
	})
}

type autosaveItem struct {
	QuestionID string `json:"question_id" binding:"required,uuid"`
	Value      string `json:"value"`
}

type autosaveRequest struct {
	Answers []autosaveItem `json:"answers" binding:"required,min=1,dive"`
}

// Autosave: simpan batch jawaban dari interval FE.
func (h *AttemptEngineHandler) Autosave(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}

	var req autosaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	items := make([]usecase.AutosaveItem, 0, len(req.Answers))
	for _, a := range req.Answers {
		qid, err := uuid.Parse(a.QuestionID)
		if err != nil {
			c.Error(apperror.BadRequest("question_id tidak valid: "+a.QuestionID, err))
			return
		}
		items = append(items, usecase.AutosaveItem{QuestionID: qid, Value: a.Value})
	}

	saved, err := h.uc.Autosave(c.Request.Context(), userID, attemptID, items)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Autosave berhasil", ginH{
		"saved_count": saved,
	})
}

type submitAttemptRequest struct {
	ConfirmSubmit bool `json:"confirm_submit"`
}

// Submit: finalisasi attempt (manual atau auto-submit saat waktu habis).
func (h *AttemptEngineHandler) Submit(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}

	var req submitAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	attempt, err := h.uc.Submit(c.Request.Context(), userID, attemptID, req.ConfirmSubmit)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Ujian dikumpulkan", attemptResponse(attempt))
}
