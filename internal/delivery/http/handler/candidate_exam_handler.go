package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/repository"
	mw "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type CandidateExamHandler struct {
	uc          *usecase.CandidateExamUsecase
	attemptEng  *usecase.AttemptEngineUsecase
	reportUC    *usecase.QuestionReportUsecase
	gradingRepo *repository.GradingRepository
}

func NewCandidateExamHandler(
	uc *usecase.CandidateExamUsecase,
	attemptEng *usecase.AttemptEngineUsecase,
	reportUC *usecase.QuestionReportUsecase,
	gradingRepo *repository.GradingRepository,
) *CandidateExamHandler {
	return &CandidateExamHandler{uc: uc, attemptEng: attemptEng, reportUC: reportUC, gradingRepo: gradingRepo}
}

func (h *CandidateExamHandler) principalID(c *gin.Context) (uuid.UUID, bool) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(principal.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
		return uuid.Nil, false
	}
	return id, true
}

func (h *CandidateExamHandler) ListExams(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	rows, err := h.uc.ListExams(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Daftar ujian", rows)
}

type validateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func bindToken(c *gin.Context) (string, bool) {
	var req validateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return "", false
	}
	return req.Token, true
}

func (h *CandidateExamHandler) ValidateToken(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	examID, err := uuid.Parse(c.Param("exam_id"))
	if err != nil {
		c.Error(apperror.BadRequest("exam_id tidak valid", err))
		return
	}
	token, ok := bindToken(c)
	if !ok {
		return
	}
	if err := h.uc.ValidateToken(c.Request.Context(), userID, examID, token); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Token valid", gin.H{"valid": true})
}

func (h *CandidateExamHandler) Start(c *gin.Context) {
	userID, ok := h.principalID(c)
	if !ok {
		return
	}
	examID, err := uuid.Parse(c.Param("exam_id"))
	if err != nil {
		c.Error(apperror.BadRequest("exam_id tidak valid", err))
		return
	}

	var req validateTokenRequest
	_ = c.ShouldBindJSON(&req)

	attempt, err := h.uc.Start(c.Request.Context(), userID, examID, req.Token)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Attempt dimulai", ginH{
		"attempt_id": attempt.ID,
		"started_at": attempt.StartedAt,
		"expires_at": attempt.ExpiresAt,
		"attempt_no": attempt.AttemptNo,
	})
}

// GetDiscussion: pembahasan soal setelah submit (kunci + pembahasan + skor).
func (h *CandidateExamHandler) GetDiscussion(c *gin.Context) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}
	userID, err := uuid.Parse(principal.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
		return
	}
	attemptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID attempt tidak valid", err))
		return
	}

	_, items, err := h.attemptEng.GetDiscussion(c.Request.Context(), userID, attemptID)
	if err != nil {
		c.Error(err)
		return
	}

	discussion := make([]gin.H, 0, len(items))
	for _, item := range items {
		opts := make([]gin.H, 0, len(item.Options))
		for _, o := range item.Options {
			opt := gin.H{"option_key": o.OptionKey, "text": o.Text}
			if o.Media != nil {
				opt["media_url"] = "/api/v1/media/" + o.Media.ID.String() + "/file"
			}
			opts = append(opts, opt)
		}
		media := gin.H(nil)
		if item.Media != nil {
			media = ginH{
				"url":      "/api/v1/media/" + item.Media.ID.String() + "/file",
				"position": item.MediaPosition,
			}
		}
		discussion = append(discussion, gin.H{
			"question_id":  item.QuestionID,
			"section_name": item.SectionName,
			"type":         item.Type,
			"text":         item.Text,
			"score_weight": item.ScoreWeight,
			"media":        media,
			"options":      opts,
			"correct_keys": item.CorrectKeys,
			"explanation":  item.Explanation,
			"answer_value": item.AnswerValue,
			"is_correct":   item.IsCorrect,
			"score":        item.Score,
			"feedback":     item.Feedback,
			"is_flagged":   item.IsFlagged,
		})
	}
	response.Success(c, http.StatusOK, "Pembahasan soal", discussion)
}

type reportQuestionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// ReportQuestion: siswa laporkan soal bermasalah.
func (h *CandidateExamHandler) ReportQuestion(c *gin.Context) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}
	userID, err := uuid.Parse(principal.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
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

	var req reportQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportUC.Create(c.Request.Context(), usecase.CreateReportInput{
		AttemptID:  attemptID,
		QuestionID: questionID,
		UserID:     userID,
		Reason:     req.Reason,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Laporan soal terkirim", report)
}
