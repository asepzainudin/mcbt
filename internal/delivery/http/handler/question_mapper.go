package handler

import (
	"strings"

	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

// normalizeQuestionType accepts spec UPPERCASE and legacy lowercase forms.
func normalizeQuestionType(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func optionKey(label string) string {
	return strings.ToUpper(strings.TrimSpace(label))
}

// bankResponse maps a QuestionBank to the API envelope (spec fields included).
func bankResponse(b *model.QuestionBank) ginH {
	return ginH{
		"id":               b.ID,
		"code":             b.Code,
		"title":            b.Title,
		"subject_id":       b.SubjectID,
		"academic_year_id": b.AcademicYearID,
		"status":           b.Status,
		"description":      b.Description,
		"subject":          b.Subject,
		"academic_year":    b.AcademicYear,
		"created_at":       b.CreatedAt,
		"updated_at":       b.UpdatedAt,
	}
}

// questionResponse maps a Question to the spec envelope: type (upper),
// text/score_weight aliases plus option_key on every option.
func questionResponse(q *model.Question) ginH {
	opts := make([]ginH, 0, len(q.Options))
	for _, o := range q.Options {
		opts = append(opts, ginH{
			"id":         o.ID,
			"option_key": optionKey(o.Label),
			"label":      o.Label,
			"text":       o.Content,
			"content":    o.Content,
			"is_correct": o.IsCorrect,
			"position":   o.Position,
			"media_id":   o.MediaID,
			"media":      o.Media,
		})
	}

	resp := ginH{
		"id":               q.ID,
		"question_bank_id": q.QuestionBankID,
		"type":             strings.ToUpper(q.QuestionType),
		"question_type":    q.QuestionType,
		"text":             q.Content,
		"content":          q.Content,
		"score_weight":     q.ScoreWeight,
		"points":           q.ScoreWeight,
		"explanation":      q.Explanation,
		"answer_keys":      answerKeysList(q.AnswerKeys),
		"media_id":         q.MediaID,
		"media":            q.Media,
		"options":          opts,
		"created_at":       q.CreatedAt,
		"updated_at":       q.UpdatedAt,
	}
	return resp
}

func answerKeysList(raw *string) []string {
	if raw == nil || *raw == "" {
		return []string{}
	}
	keys := strings.Split(*raw, "\n")
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" {
			out = append(out, k)
		}
	}
	return out
}

func joinAnswerKeys(keys []string) *string {
	cleaned := make([]string, 0, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" {
			cleaned = append(cleaned, k)
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	joined := strings.Join(cleaned, "\n")
	return &joined
}

type ginH = map[string]any

func toUUIDPtr(raw *string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, apperror.BadRequest("UUID tidak valid", err)
	}
	return &id, nil
}
