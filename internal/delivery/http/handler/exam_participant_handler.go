package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type ExamParticipantHandler struct {
	uc *usecase.ExamParticipantUsecase
}

func NewExamParticipantHandler(uc *usecase.ExamParticipantUsecase) *ExamParticipantHandler {
	return &ExamParticipantHandler{uc: uc}
}

func participantResponse(p *model.ExamParticipant) ginH {
	resp := ginH{
		"id":           p.ID,
		"exam_id":      p.ExamID,
		"student_id":   p.StudentID,
		"assigned_via": p.AssignedVia,
		"nis":          "",
		"name":         "",
		"class_name":   nil,
		"created_at":   p.CreatedAt,
	}
	if p.Student != nil {
		resp["nis"] = p.Student.Nis
		if p.Student.User != nil {
			resp["name"] = p.Student.User.Name
		}
		if p.Student.Class != nil {
			resp["class_name"] = p.Student.Class.Name
		}
	}
	return resp
}

func (h *ExamParticipantHandler) List(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	participants, err := h.uc.List(c.Request.Context(), examID)
	if err != nil {
		c.Error(err)
		return
	}
	data := make([]ginH, 0, len(participants))
	for i := range participants {
		data = append(data, participantResponse(&participants[i]))
	}
	response.Success(c, http.StatusOK, "Peserta ditemukan", data)
}

type assignClassRequest struct {
	ClassIDs []string `json:"class_ids" binding:"required,min=1,dive,uuid"`
}

func (h *ExamParticipantHandler) AssignClass(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	var req assignClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	ids := make([]uuid.UUID, 0, len(req.ClassIDs))
	for _, raw := range req.ClassIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("class_ids tidak valid: "+raw, err))
			return
		}
		ids = append(ids, id)
	}
	result, err := h.uc.AssignClasses(c.Request.Context(), examID, ids)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Peserta dari kelas berhasil ditambahkan", result)
}

type assignIndividualRequest struct {
	StudentIDs []string `json:"student_ids" binding:"required,min=1,dive,uuid"`
}

func (h *ExamParticipantHandler) AssignIndividual(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	var req assignIndividualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	ids := make([]uuid.UUID, 0, len(req.StudentIDs))
	for _, raw := range req.StudentIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("student_ids tidak valid: "+raw, err))
			return
		}
		ids = append(ids, id)
	}
	result, err := h.uc.AssignIndividuals(c.Request.Context(), examID, ids)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Peserta berhasil ditambahkan", result)
}

func (h *ExamParticipantHandler) Remove(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	participantID, err := uuid.Parse(c.Param("participant_id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID peserta tidak valid", err))
		return
	}
	if err := h.uc.Remove(c.Request.Context(), examID, participantID); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Peserta dihapus", nil)
}
