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

type QuestionHandler struct {
	uc *usecase.QuestionUsecase
}

func NewQuestionHandler(uc *usecase.QuestionUsecase) *QuestionHandler {
	return &QuestionHandler{uc: uc}
}

func (h *QuestionHandler) List(c *gin.Context) {
	page := clampInt(c.Query("page"), 1, 1, 1_000_000)
	limit := clampInt(c.Query("limit"), 20, 1, 100)

	params := repository.QuestionListParams{
		Search: c.Query("search"),
		Type:   normalizeQuestionType(c.Query("type")),
		Page:   page,
		Limit:  limit,
	}
	if raw := c.Query("bank_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("bank_id harus UUID yang valid", err))
			return
		}
		params.BankID = &id
	}

	items, total, err := h.uc.List(c.Request.Context(), params)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]ginH, 0, len(items))
	for i := range items {
		data = append(data, questionResponse(&items[i]))
	}
	response.SuccessWithMeta(c, http.StatusOK, "Soal ditemukan", data, paginationMeta(page, limit, total))
}

func (h *QuestionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	q, err := h.uc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Detail soal", questionResponse(q))
}

// Preview returns the question rendered as a student would see it.
func (h *QuestionHandler) Preview(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	q, err := h.uc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	opts := make([]ginH, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, ginH{
			"option_key": optionKey(o.Label),
			"text":       o.Content,
			"image_url":  o.Media != nil && o.Media.FilePath != "",
			"media":      o.Media,
		})
	}

	response.Success(c, http.StatusOK, "Preview soal", ginH{
		"type":         normalizeQuestionType(q.QuestionType),
		"content_html": q.Content,
		"image":        q.Media,
		"score_weight": q.ScoreWeight,
		"options":      opts,
		"answer_keys":  answerKeysList(q.AnswerKeys),
		// kunci jawaban ikut karena preview hanya untuk admin/penulis soal
		"correct_keys": correctKeysOf(q),
		"explanation":  q.Explanation,
	})
}

func correctKeysOf(q *model.Question) []string {
	keys := []string{}
	for _, o := range q.Options {
		if o.IsCorrect {
			keys = append(keys, optionKey(o.Label))
		}
	}
	return keys
}

// questionRequest accepts BOTH the STEP-07 spec field names and the legacy ones.
type optionPayload struct {
	OptionKey string  `json:"option_key"`
	Label     string  `json:"label"`
	Text      string  `json:"text"`
	Content   string  `json:"content"`
	MediaID   *string `json:"media_id" binding:"omitempty,uuid"`
	IsCorrect bool    `json:"is_correct"`
}

type questionRequest struct {
	Type         string          `json:"type"`
	QuestionType string          `json:"question_type"`
	Text         string          `json:"text"`
	Content      string          `json:"content"`
	ScoreWeight  *float64        `json:"score_weight"`
	Points       *float64        `json:"points"`
	Explanation  *string         `json:"explanation" binding:"omitempty,max=2000"`
	MediaID      *string         `json:"media_id" binding:"omitempty,uuid"`
	BankID       string          `json:"question_bank_id" binding:"omitempty,uuid"`
	AnswerKeys   []string        `json:"answer_keys"`
	AnswerKey    string          `json:"answer_key"`
	Options      []optionPayload `json:"options"`
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (h *QuestionHandler) buildInput(c *gin.Context, bankIDFromPath *uuid.UUID) (usecase.QuestionInput, bool) {
	var req questionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return usecase.QuestionInput{}, false
	}

	input := usecase.QuestionInput{
		Type:    normalizeQuestionType(firstNonEmpty(req.Type, req.QuestionType)),
		Content: firstNonEmpty(req.Text, req.Content),
		Options: make([]usecase.OptionInput, 0, len(req.Options)),
	}

	if req.ScoreWeight != nil {
		input.ScoreWeight = *req.ScoreWeight
	} else if req.Points != nil {
		input.ScoreWeight = *req.Points
	}

	rawBank := firstNonEmpty(req.BankID, derefUUID(bankIDFromPath))
	if rawBank != "" {
		id, err := uuid.Parse(rawBank)
		if err != nil {
			c.Error(apperror.BadRequest("question_bank_id tidak valid", err))
			return usecase.QuestionInput{}, false
		}
		input.BankID = id
	}

	mediaID, err := toUUIDPtr(req.MediaID)
	if err != nil {
		return usecase.QuestionInput{}, false
	}
	input.MediaID = mediaID

	for _, o := range req.Options {
		oMediaID, err := toUUIDPtr(o.MediaID)
		if err != nil {
			return usecase.QuestionInput{}, false
		}
		input.Options = append(input.Options, usecase.OptionInput{
			Content:   firstNonEmpty(o.Text, o.Content),
			MediaID:   oMediaID,
			IsCorrect: o.IsCorrect,
		})
	}

	input.AnswerKeys = req.AnswerKeys
	if req.AnswerKey != "" {
		input.AnswerKeys = append(input.AnswerKeys, req.AnswerKey)
	}

	return input, true
}

func derefUUID(s *uuid.UUID) string {
	if s == nil {
		return ""
	}
	return s.String()
}

func (h *QuestionHandler) Create(c *gin.Context) {
	in, ok := h.buildInput(c, nil)
	if !ok {
		return
	}
	q, err := h.uc.Create(c.Request.Context(), in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Soal dibuat", questionResponse(q))
}

// CreateInBank implements POST /question-banks/{bank_id}/questions (spec route).
func (h *QuestionHandler) CreateInBank(c *gin.Context) {
	bankID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("bank_id tidak valid", err))
		return
	}
	in, ok := h.buildInput(c, &bankID)
	if !ok {
		return
	}
	in.BankID = bankID

	q, err := h.uc.Create(c.Request.Context(), in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Soal dibuat", questionResponse(q))
}

func (h *QuestionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	in, ok := h.buildInput(c, nil)
	if !ok {
		return
	}
	q, err := h.uc.Update(c.Request.Context(), id, in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Soal diperbarui", questionResponse(q))
}

func (h *QuestionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Soal dihapus", nil)
}

type reorderOptionsRequest struct {
	OptionIDsOrder []string `json:"option_ids_order" binding:"required,min=2,dive,uuid"`
}

func (h *QuestionHandler) ReorderOptions(c *gin.Context) {
	questionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID soal tidak valid", err))
		return
	}

	var req reorderOptionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	ordered := make([]uuid.UUID, 0, len(req.OptionIDsOrder))
	for _, raw := range req.OptionIDsOrder {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("ID opsi tidak valid: "+raw, err))
			return
		}
		ordered = append(ordered, id)
	}

	if err := h.uc.ReorderOptions(c.Request.Context(), questionID, ordered); err != nil {
		c.Error(err)
		return
	}
	q, _ := h.uc.Get(c.Request.Context(), questionID)
	response.Success(c, http.StatusOK, "Urutan opsi diperbarui", questionResponse(q))
}

type optionUpdateRequest struct {
	IsCorrect bool `json:"is_correct"`
}

func (h *QuestionHandler) UpdateOption(c *gin.Context) {
	questionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID soal tidak valid", err))
		return
	}
	optionID, err := uuid.Parse(c.Param("option_id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID opsi tidak valid", err))
		return
	}

	var req optionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if !req.IsCorrect {
		c.Error(apperror.New(http.StatusUnprocessableEntity,
			"Sertakan is_correct=true untuk menandai jawaban benar", nil))
		return
	}

	q, err := h.uc.SetCorrectOption(c.Request.Context(), questionID, optionID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Jawaban benar ditandai", questionResponse(q))
}
