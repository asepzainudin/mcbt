package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type QuestionBankHandler struct {
	uc *usecase.QuestionBankUsecase
}

func NewQuestionBankHandler(uc *usecase.QuestionBankUsecase) *QuestionBankHandler {
	return &QuestionBankHandler{uc: uc}
}

func (h *QuestionBankHandler) List(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 10, 1, 100)

	var subjectID *uuid.UUID
	if raw := c.Query("subject_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("subject_id harus UUID yang valid", err))
			return
		}
		subjectID = &id
	}

	// semua staff melihat seluruh bank; kepemilikan dicek saat mutasi
	items, total, err := h.uc.List(c.Request.Context(), c.Query("search"), subjectID, nil, page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]ginH, 0, len(items))
	for i := range items {
		data = append(data, bankResponseActor(&items[i], actorID, actorIsAdmin))
	}
	response.SuccessWithMeta(c, http.StatusOK, "Bank soal ditemukan", data, paginationMeta(page, limit, total))
}

type questionBankRequest struct {
	Code           string  `json:"code" binding:"required,min=3,max=50"`
	SubjectID      string  `json:"subject_id" binding:"required,uuid"`
	AcademicYearID *string `json:"academic_year_id" binding:"omitempty,uuid"`
	Title          string  `json:"title" binding:"required,max=150"`
	Description    *string `json:"description" binding:"omitempty,max=500"`
}

func parseOptionalUUID(raw *string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, apperror.BadRequest("academic_year_id tidak valid", err)
	}
	return &id, nil
}

func bindQuestionBankRequest(c *gin.Context) (usecase.QuestionBankInput, bool) {
	var req questionBankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return usecase.QuestionBankInput{}, false
	}
	subjectID, err := uuid.Parse(req.SubjectID)
	if err != nil {
		c.Error(apperror.BadRequest("subject_id tidak valid", err))
		return usecase.QuestionBankInput{}, false
	}
	ayID, err := parseOptionalUUID(req.AcademicYearID)
	if err != nil {
		return usecase.QuestionBankInput{}, false
	}
	return usecase.QuestionBankInput{
		SubjectID:      subjectID,
		AcademicYearID: ayID,
		Code:           req.Code,
		Title:          req.Title,
		Description:    req.Description,
	}, true
}

func (h *QuestionBankHandler) Create(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	in, ok := bindQuestionBankRequest(c)
	if !ok {
		return
	}
	if uid, _, ok2 := principalActor(c); ok2 {
		in.CreatedBy = &uid // bank dimiliki pembuatnya
	}
	qb, err := h.uc.Create(c.Request.Context(), in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Bank soal dibuat", bankResponseActor(qb, actorID, actorIsAdmin))
}

func (h *QuestionBankHandler) Update(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	in, ok := bindQuestionBankRequest(c)
	if !ok {
		return
	}
	qb, err := h.uc.Update(c.Request.Context(), id, in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Bank soal diperbarui", bankResponseActor(qb, actorID, actorIsAdmin))
}

func (h *QuestionBankHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Bank soal dihapus", nil)
}

func (h *QuestionBankHandler) Clone(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	actorID, _, ok := principalActor(c)
	if !ok {
		c.Error(apperror.New(http.StatusUnauthorized, "Authentication required", nil))
		return
	}
	qb, err := h.uc.Clone(c.Request.Context(), id, actorID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Bank soal berhasil dikloning", bankResponseActor(qb, actorID, actorIsAdmin))
}

func (h *QuestionBankHandler) Publish(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	qb, err := h.uc.Publish(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Bank soal dipublikasikan", bankResponseActor(qb, actorID, actorIsAdmin))
}

func (h *QuestionBankHandler) Archive(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	qb, err := h.uc.Archive(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Bank soal diarsipkan", bankResponseActor(qb, actorID, actorIsAdmin))
}
