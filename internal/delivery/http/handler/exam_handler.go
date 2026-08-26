package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/repository"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type ExamHandler struct {
	access *usecase.AccessUsecase

	uc *usecase.ExamUsecase
}

func NewExamHandler(uc *usecase.ExamUsecase, access *usecase.AccessUsecase) *ExamHandler {
	return &ExamHandler{uc: uc, access: access}
}

func examResponse(e *model.Exam, userID uuid.UUID, isAdmin bool) ginH {
	return ginH{
		"can_manage":              isAdmin || (e.CreatedBy != nil && *e.CreatedBy == userID),
		"id":                      e.ID,
		"title":                   e.Title,
		"description":             e.Description,
		"subject_id":              e.SubjectID,
		"academic_year_id":        e.AcademicYearID,
		"question_bank_id":        e.QuestionBankID,
		"status":                  e.Status,
		"duration_minutes":        e.DurationMinutes,
		"max_attempts":            e.MaxAttempts,
		"passing_grade":           e.PassingGrade,
		"randomize_questions":     e.RandomizeQuestions,
		"randomize_options":       e.RandomizeOptions,
		"allow_backtrack":         e.AllowBacktrack,
		"auto_submit":             e.AutoSubmit,
		"show_result_immediately": e.ShowResultImmediately,
		"allow_discussion":        e.AllowDiscussion,
		"negative_marking":        e.NegativeMarking,
		"negative_value":          e.NegativeValue,
		"token_enabled":           e.TokenEnabled,
		"attempts_count":          e.AttemptsCount,
		"exam_token":              e.ExamToken,
		"subject":                 e.Subject,
		"academic_year":           e.AcademicYear,
		"question_bank":           e.QuestionBank,
		"schedule":                e.Schedule,
		"created_at":              e.CreatedAt,
		"updated_at":              e.UpdatedAt,
	}
}

// examActor: identitas aktor utk perhitungan can_manage.
func examActor(c *gin.Context) (uuid.UUID, bool) {
	uid, isAdmin, ok := principalActor(c)
	if !ok {
		return uuid.Nil, false
	}
	return uid, isAdmin
}

func (h *ExamHandler) List(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 10, 1, 100)

	params := repository.ExamListParams{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Page:   page,
		Limit:  limit,
	}
	for _, q := range []struct {
		key string
		dst **uuid.UUID
	}{
		{"subject_id", &params.SubjectID},
		{"academic_year_id", &params.AcademicYearID},
	} {
		if raw := c.Query(q.key); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				c.Error(apperror.BadRequest(q.key+" harus UUID yang valid", err))
				return
			}
			*q.dst = &id
		}
	}

	if uid, isAdmin, ok := principalActor(c); ok && !isAdmin {
		params.OwnerUserID = &uid // guru hanya melihat ujian buatannya sendiri
	}

	items, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]ginH, 0, len(items))
	for i := range items {
		data = append(data, examResponse(&items[i], actorID, actorIsAdmin))
	}
	response.SuccessWithMeta(c, http.StatusOK, "Ujian ditemukan", data, paginationMeta(page, limit, total))
}

func (h *ExamHandler) Get(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	e, err := h.uc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Detail ujian", examResponse(e, actorID, actorIsAdmin))
}

type examRequest struct {
	Title          string  `json:"title" binding:"required,max=150"`
	Description    *string `json:"description" binding:"omitempty,max=1000"`
	SubjectID      string  `json:"subject_id" binding:"required,uuid"`
	AcademicYearID *string `json:"academic_year_id" binding:"omitempty,uuid"`
	QuestionBankID *string `json:"question_bank_id" binding:"omitempty,uuid"`
}

func (h *ExamHandler) buildInput(c *gin.Context) (usecase.ExamInput, bool) {
	var req examRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return usecase.ExamInput{}, false
	}

	subjectID, err := uuid.Parse(req.SubjectID)
	if err != nil {
		c.Error(apperror.BadRequest("subject_id tidak valid", err))
		return usecase.ExamInput{}, false
	}
	ayID, err := parseOptionalUUID(req.AcademicYearID)
	if err != nil {
		return usecase.ExamInput{}, false
	}
	bankID, err := parseOptionalUUID(req.QuestionBankID)
	if err != nil {
		return usecase.ExamInput{}, false
	}

	return usecase.ExamInput{
		Title:          req.Title,
		Description:    req.Description,
		SubjectID:      subjectID,
		AcademicYearID: ayID,
		QuestionBankID: bankID,
	}, true
}

func (h *ExamHandler) Create(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	in, ok := h.buildInput(c)
	if !ok {
		return
	}
	if actorID != uuid.Nil {
		in.CreatedBy = &actorID // ujian dimiliki pembuatnya
	}
	e, err := h.uc.Create(c.Request.Context(), in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Ujian dibuat", examResponse(e, actorID, actorIsAdmin))
}

func (h *ExamHandler) Update(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	in, ok := h.buildInput(c)
	if !ok {
		return
	}
	e, err := h.uc.Update(c.Request.Context(), id, in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Ujian diperbarui", examResponse(e, actorID, actorIsAdmin))
}

type examSettingsRequest struct {
	DurationMinutes       int     `json:"duration_minutes"`
	MaxAttempts           int     `json:"max_attempts"`
	PassingGrade          float64 `json:"passing_grade"`
	RandomizeQuestions    bool    `json:"randomize_questions"`
	RandomizeOptions      bool    `json:"randomize_options"`
	AllowBacktrack        *bool   `json:"allow_backtrack"`
	AutoSubmit            *bool   `json:"auto_submit"`
	ShowResultImmediately bool    `json:"show_result_immediately"`
	NegativeMarking       bool    `json:"negative_marking"`
	NegativeValue         float64 `json:"negative_value"`
	TokenEnabled          bool    `json:"token_enabled"`
	AllowDiscussion       bool    `json:"allow_discussion"`
}

func (h *ExamHandler) UpdateSettings(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}

	var req examSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	in := usecase.ExamSettingsInput{
		DurationMinutes:       req.DurationMinutes,
		MaxAttempts:           req.MaxAttempts,
		PassingGrade:          req.PassingGrade,
		RandomizeQuestions:    req.RandomizeQuestions,
		RandomizeOptions:      req.RandomizeOptions,
		AutoSubmit:            true,
		ShowResultImmediately: req.ShowResultImmediately,
		NegativeMarking:       req.NegativeMarking,
		NegativeValue:         req.NegativeValue,
		TokenEnabled:          req.TokenEnabled,
		AllowDiscussion:       req.AllowDiscussion,
	}
	if req.AllowBacktrack != nil {
		in.AllowBacktrack = *req.AllowBacktrack
	} else {
		in.AllowBacktrack = true
	}
	if req.AutoSubmit != nil {
		in.AutoSubmit = *req.AutoSubmit
	}

	e, err := h.uc.UpdateSettings(c.Request.Context(), id, in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Pengaturan ujian disimpan", examResponse(e, actorID, actorIsAdmin))
}

func (h *ExamHandler) Publish(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	e, err := h.uc.Publish(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Ujian dipublikasikan", examResponse(e, actorID, actorIsAdmin))
}

func (h *ExamHandler) Close(c *gin.Context) {
	actorID, actorIsAdmin := examActor(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	e, err := h.uc.Close(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Ujian ditutup", examResponse(e, actorID, actorIsAdmin))
}

func (h *ExamHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Ujian dihapus", nil)
}
