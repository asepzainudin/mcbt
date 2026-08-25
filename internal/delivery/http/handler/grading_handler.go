package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type GradingHandler struct {
	uc *usecase.GradingUsecase
}

func NewGradingHandler(uc *usecase.GradingUsecase) *GradingHandler {
	return &GradingHandler{uc: uc}
}

// CalculateGrades: engine penilaian otomatis untuk seluruh attempt submitted.
func (h *GradingHandler) CalculateGrades(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	result, err := h.uc.CalculateGrades(c.Request.Context(), examID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Nilai otomatis selesai diproses", result)
}

func (h *GradingHandler) UngradedEssays(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	rows, err := h.uc.UngradedEssays(c.Request.Context(), examID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Esai belum dinilai", rows)
}

type gradeEssayRequest struct {
	QuestionID string  `json:"question_id" binding:"required,uuid"`
	Score      float64 `json:"score" binding:"required,gte=0"`
	Feedback   *string `json:"feedback" binding:"omitempty,max=1000"`
}

// GradeEssay: koreksi manual satu jawaban esai pada sebuah attempt.
func (h *GradingHandler) GradeEssay(c *gin.Context) {
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}
	var req gradeEssayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		c.Error(apperror.BadRequest("question_id tidak valid", err))
		return
	}

	answer, err := h.uc.GradeEssay(c.Request.Context(), attemptID, questionID, usecase.GradeEssayInput{
		QuestionID: questionID,
		Score:      req.Score,
		Feedback:   req.Feedback,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Nilai esai tersimpan", ginH{
		"question_id": answer.QuestionID,
		"score":       answer.Score,
		"feedback":    answer.Feedback,
		"graded_via":  answer.GradedVia,
	})
}
