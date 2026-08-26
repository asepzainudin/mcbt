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

type QuestionReportHandler struct {
	uc     *usecase.QuestionReportUsecase
	access *usecase.AccessUsecase
}

func NewQuestionReportHandler(uc *usecase.QuestionReportUsecase, access *usecase.AccessUsecase) *QuestionReportHandler {
	return &QuestionReportHandler{uc: uc, access: access}
}

type createReportRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// Create: siswa laporkan soal bermasalah.
func (h *QuestionReportHandler) Create(c *gin.Context) {
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

	var req createReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	report, err := h.uc.Create(c.Request.Context(), usecase.CreateReportInput{
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

// List: admin lihat semua laporan (opsional filter ?status=).
func (h *QuestionReportHandler) List(c *gin.Context) {
	var ownerUserID *uuid.UUID
	if uid, isAdmin, ok := principalActor(c); ok && !isAdmin {
		ownerUserID = &uid // guru hanya melihat laporan pada ujiannya
	}
	reports, err := h.uc.List(c.Request.Context(), c.Query("status"), ownerUserID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Laporan soal", reports)
}

type resolveReportRequest struct {
	Status     string  `json:"status" binding:"required,oneof=pending reviewing resolved rejected"`
	Resolution *string `json:"resolution"`
}

// Resolve: admin tangani laporan.
func (h *QuestionReportHandler) Resolve(c *gin.Context) {
	principal, ok := mw.CurrentPrincipal(c)
	if !ok {
		return
	}
	resolverID, err := uuid.Parse(principal.UserID)
	if err != nil {
		c.Error(apperror.BadRequest("Invalid user id", err))
		return
	}
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID laporan tidak valid", err))
		return
	}

	var req resolveReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	report, err := h.uc.Resolve(c.Request.Context(), reportID, resolverID, usecase.ResolveInput{
		Status:     req.Status,
		Resolution: req.Resolution,
	}, principal.HasRole("admin"))
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Laporan ditangani", report)
}
