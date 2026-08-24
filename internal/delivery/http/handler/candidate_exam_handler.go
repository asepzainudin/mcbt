package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	mw "github.com/asepzainudin14/mcbt/internal/server/middleware"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type CandidateExamHandler struct {
	uc *usecase.CandidateExamUsecase
}

func NewCandidateExamHandler(uc *usecase.CandidateExamUsecase) *CandidateExamHandler {
	return &CandidateExamHandler{uc: uc}
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
	_ = c.ShouldBindJSON(&req) // token opsional utk ujian tanpa token protection

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
