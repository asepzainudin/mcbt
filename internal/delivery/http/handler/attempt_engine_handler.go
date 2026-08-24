package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	mw "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type AttemptEngineHandler struct {
	uc *usecase.AttemptEngineUsecase
}

func NewAttemptEngineHandler(uc *usecase.AttemptEngineUsecase) *AttemptEngineHandler {
	return &AttemptEngineHandler{uc: uc}
}

func (h *AttemptEngineHandler) principalID(c *gin.Context) (uuid.UUID, bool) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return uuid.Nil, false
	}
	id, err := uuid.Parse(principal.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
		return uuid.Nil, false
	}
	return id, true
}

func attemptResponse(a *model.ExamAttempt) ginH {
	return ginH{
		"attempt_id":   a.ID,
		"exam_id":      a.ExamID,
		"attempt_no":   a.AttemptNo,
		"status":       a.Status,
		"started_at":   a.StartedAt,
		"expires_at":   a.ExpiresAt,
		"submitted_at": a.SubmittedAt,
	}
}

func attemptQuestionResponse(item usecase.AttemptQuestionItem) ginH {
	opts := make([]ginH, 0, len(item.Options))
	for _, o := range item.Options {
		opts = append(opts, ginH{
			"option_key": o.OptionKey,
			"text":       o.Text,
			"media":      mediaRefResponse(o.Media),
		})
	}
	return ginH{
		"question_id":    item.QuestionID,
		"section_name":   item.SectionName,
		"sequence":       item.Sequence,
		"type":           item.Type,
		"text":           item.Text,
		"score_weight":   item.ScoreWeight,
		"media":          mediaRefResponse(item.Media),
		"media_position": item.MediaPosition,
		"options":        opts,
		"answer_value":   item.AnswerValue,
		"is_flagged":     item.IsFlagged,
		"answered_at":    item.AnsweredAt,
	}
}

// GetQuestions: lembar soal + jawaban tersimpan + status flag.
func (h *AttemptEngineHandler) GetQuestions(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}

	attempt, items, err := h.uc.GetQuestions(c.Request.Context(), userID, attemptID)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]ginH, 0, len(items))
	for _, item := range items {
		data = append(data, attemptQuestionResponse(item))
	}
	response.Success(c, http.StatusOK, "Lembar soal", ginH{
		"attempt":  attemptResponse(attempt),
		"questions": data,
	})
}

type saveAnswerRequest struct {
	QuestionID      string `json:"question_id" binding:"required,uuid"`
	AnswerValue     string `json:"answer_value"`
	ClientTimestamp int64  `json:"client_timestamp"`
}

// SaveAnswer: simpan jawaban real-time (upsert).
func (h *AttemptEngineHandler) SaveAnswer(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}

	var req saveAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		c.Error(apperror.BadRequest("question_id tidak valid", err))
		return
	}

	answer, err := h.uc.SaveAnswer(c.Request.Context(), userID, attemptID, usecase.SaveAnswerInput{
		QuestionID:      questionID,
		AnswerValue:     req.AnswerValue,
		ClientTimestamp: req.ClientTimestamp,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Jawaban tersimpan", ginH{
		"question_id":      answer.QuestionID,
		"answer_value":     answer.AnswerValue,
		"is_flagged":       answer.IsFlagged,
		"answered_at":      answer.AnsweredAt,
		"client_timestamp": answer.ClientTimestamp,
	})
}

// Flag: tandai ragu-ragu.
func (h *AttemptEngineHandler) Flag(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}
	questionID, err := uuid.Parse(c.Param("question_id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID soal tidak valid", err))
		return
	}

	answer, err := h.uc.SetFlag(c.Request.Context(), userID, attemptID, questionID, true)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Soal ditandai ragu-ragu", ginH{
		"question_id": answer.QuestionID,
		"is_flagged":  answer.IsFlagged,
	})
}

// Unflag: lepas tandai ragu-ragu.
func (h *AttemptEngineHandler) Unflag(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}
	questionID, err := uuid.Parse(c.Param("question_id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID soal tidak valid", err))
		return
	}

	answer, err := h.uc.SetFlag(c.Request.Context(), userID, attemptID, questionID, false)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Tanda ragu-ragu dilepas", ginH{
		"question_id": answer.QuestionID,
		"is_flagged":  answer.IsFlagged,
	})
}
