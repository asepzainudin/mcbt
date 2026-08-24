package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/response"
	"github.com/asepzainudin14/mcbt/internal/usecase"
)

type ExamSectionHandler struct {
	uc *usecase.ExamSectionUsecase
}

func NewExamSectionHandler(uc *usecase.ExamSectionUsecase) *ExamSectionHandler {
	return &ExamSectionHandler{uc: uc}
}

func sectionResponse(s *model.ExamSection, questionCount int64) ginH {
	return ginH{
		"id":             s.ID,
		"exam_id":        s.ExamID,
		"name":           s.Name,
		"sequence":       s.Sequence,
		"question_count": questionCount,
		"created_at":     s.CreatedAt,
		"updated_at":     s.UpdatedAt,
	}
}

func (h *ExamSectionHandler) ListByExam(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}

	sections, err := h.uc.ListByExam(c.Request.Context(), examID)
	if err != nil {
		c.Error(err)
		return
	}

	data := make([]ginH, 0, len(sections))
	for i := range sections {
		data = append(data, sectionResponse(&sections[i].ExamSection, sections[i].QuestionCount))
	}
	response.Success(c, http.StatusOK, "Sections ditemukan", data)
}

type examSectionRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	Sequence int    `json:"sequence" binding:"required,min=1"`
}

func bindSectionRequest(c *gin.Context) (usecase.ExamSectionInput, bool) {
	var req examSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return usecase.ExamSectionInput{}, false
	}
	return usecase.ExamSectionInput{Name: req.Name, Sequence: req.Sequence}, true
}

func (h *ExamSectionHandler) Create(c *gin.Context) {
	examID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID ujian tidak valid", err))
		return
	}
	in, ok := bindSectionRequest(c)
	if !ok {
		return
	}
	section, err := h.uc.Create(c.Request.Context(), examID, in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusCreated, "Section dibuat", sectionResponse(section, 0))
}

func (h *ExamSectionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	in, ok := bindSectionRequest(c)
	if !ok {
		return
	}
	section, err := h.uc.Update(c.Request.Context(), id, in)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Section diperbarui", sectionResponse(section, 0))
}

func (h *ExamSectionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID tidak valid", err))
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Section dihapus", nil)
}

type mapQuestionsRequest struct {
	QuestionBankIDs      []string `json:"question_bank_ids" binding:"required,min=1,dive,uuid"`
	TotalRandomQuestions int      `json:"total_random_questions"`
}

func (h *ExamSectionHandler) MapQuestions(c *gin.Context) {
	sectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID section tidak valid", err))
		return
	}

	var req mapQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	bankIDs := make([]uuid.UUID, 0, len(req.QuestionBankIDs))
	for _, raw := range req.QuestionBankIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.Error(apperror.BadRequest("question_bank_ids tidak valid: "+raw, err))
			return
		}
		bankIDs = append(bankIDs, id)
	}

	mapped, skipped, err := h.uc.MapQuestions(c.Request.Context(), sectionID, usecase.MapQuestionsInput{
		BankIDs:              bankIDs,
		TotalRandomQuestions: req.TotalRandomQuestions,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Mapping soal selesai", gin.H{
		"mapped_count": mapped,
		"skipped":      skipped,
	})
}

func (h *ExamSectionHandler) ListQuestions(c *gin.Context) {
	sectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID section tidak valid", err))
		return
	}
	questions, err := h.uc.ListQuestions(c.Request.Context(), sectionID)
	if err != nil {
		c.Error(err)
		return
	}
	data := make([]ginH, 0, len(questions))
	for i := range questions {
		q := &questions[i]
		data = append(data, ginH{
			"id":               q.ID,
			"type":             strings.ToUpper(q.QuestionType),
			"text":             q.Content,
			"score_weight":     q.ScoreWeight,
			"answer_keys":      answerKeysList(q.AnswerKeys),
			"question_bank_id": q.QuestionBankID,
		})
	}
	response.Success(c, http.StatusOK, "Soal dalam section", data)
}

func (h *ExamSectionHandler) RemoveQuestion(c *gin.Context) {
	sectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID section tidak valid", err))
		return
	}
	questionID, err := uuid.Parse(c.Param("question_id"))
	if err != nil {
		c.Error(apperror.BadRequest("ID soal tidak valid", err))
		return
	}

	if err := h.uc.RemoveQuestion(c.Request.Context(), sectionID, questionID); err != nil {
		c.Error(err)
		return
	}
	response.Success(c, http.StatusOK, "Soal dikeluarkan dari section", nil)
}
