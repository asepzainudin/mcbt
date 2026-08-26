package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type GradingUsecase struct {
	grading  GradingRepo
	answers  ExamAnswerRepo
	attempts ExamAttemptRepo
	exams    ExamRepo
}

func NewGradingUsecase(
	grading GradingRepo,
	answers ExamAnswerRepo,
	attempts ExamAttemptRepo,
	exams ExamRepo,
) *GradingUsecase {
	return &GradingUsecase{grading: grading, answers: answers, attempts: attempts, exams: exams}
}

// gradeObjective menilai satu jawaban objektif. Mengembalikan (score, isCorrect, false) —
// false berarti tipe soal tidak dinilai otomatis (essay).
func gradeObjective(q *model.Question, answerValue string, negativeValue float64) (float64, *bool, bool) {
	switch q.QuestionType {
	case model.QuestionTypeMultipleChoice, model.QuestionTypeTrueFalse:
		var correctKey string
		for _, o := range q.Options {
			if o.IsCorrect {
				correctKey = o.Label
				break
			}
		}
		isCorrect := strings.EqualFold(strings.TrimSpace(answerValue), correctKey)
		score := 0.0
		if isCorrect {
			score = q.ScoreWeight
		} else if negativeValue > 0 && strings.TrimSpace(answerValue) != "" {
			score = -negativeValue
		}
		return score, &isCorrect, true

	case model.QuestionTypeMultipleAnswer:
		correct := map[string]bool{}
		for _, o := range q.Options {
			if o.IsCorrect {
				correct[strings.ToUpper(strings.TrimSpace(o.Label))] = true
			}
		}
		given := map[string]bool{}
		for _, k := range strings.Split(strings.ToUpper(answerValue), ",") {
			k = strings.TrimSpace(k)
			if k != "" {
				given[k] = true
			}
		}
		isCorrect := len(given) > 0
		for k := range given {
			if !correct[k] {
				isCorrect = false
				break
			}
		}
		for k := range correct {
			if !given[k] {
				isCorrect = false
				break
			}
		}
		score := 0.0
		if isCorrect {
			score = q.ScoreWeight
		} else if negativeValue > 0 && strings.TrimSpace(answerValue) != "" {
			score = -negativeValue
		}
		return score, &isCorrect, true

	case model.QuestionTypeShortAnswer:
		var keys []string
		if q.AnswerKeys != nil {
			for _, k := range strings.Split(*q.AnswerKeys, "\n") {
				k = strings.TrimSpace(k)
				if k != "" {
					keys = append(keys, strings.ToLower(k))
				}
			}
		}
		given := strings.ToLower(strings.TrimSpace(answerValue))
		isCorrect := false
		for _, k := range keys {
			if k == given {
				isCorrect = true
				break
			}
		}
		score := 0.0
		if isCorrect {
			score = q.ScoreWeight
		} else if negativeValue > 0 && strings.TrimSpace(answerValue) != "" {
			score = -negativeValue
		}
		return score, &isCorrect, true
	}
	return 0, nil, false // essay
}

type CalculateGradesResult struct {
	AttemptsGraded  int `json:"attempts_graded"`
	QuestionsGraded int `json:"questions_graded"`
}

func (u *GradingUsecase) CalculateGrades(ctx context.Context, examID uuid.UUID) (*CalculateGradesResult, error) {
	exam, err := u.exams.FindByID(ctx, examID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	attempts, err := u.grading.ListSubmittedAttempts(ctx, exam.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	result := &CalculateGradesResult{}
	negative := 0.0
	if exam.NegativeMarking {
		negative = exam.NegativeValue
	}

	for _, attempt := range attempts {
		answers, err := u.grading.ListAnswersByAttempt(ctx, attempt.ID)
		if err != nil {
			return nil, apperror.Internal(err)
		}

		for _, a := range answers {
			q, err := u.grading.QuestionWithGradingInfo(ctx, a.QuestionID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return nil, apperror.Internal(err)
			}

			score, isCorrect, auto := gradeObjective(q, a.AnswerValue, negative)
			if !auto {
				continue // essay: menunggu koreksi manual
			}
			via := "auto"
			if err := u.grading.UpdateGrading(ctx, a.ID, score, isCorrect, nil, via); err != nil {
				return nil, err
			}
			result.QuestionsGraded++
		}

		// total dihitung ulang dari DB: nilai objektif baru + nilai esai manual tetap utuh
		total, err := u.grading.SumScoresByAttempt(ctx, attempt.ID)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if err := u.grading.UpdateAttemptScore(ctx, attempt.ID, total); err != nil {
			return nil, err
		}
		result.AttemptsGraded++
	}
	return result, nil
}

type UngradedEssayRow = repository.UngradedEssayRow

func (u *GradingUsecase) UngradedEssays(ctx context.Context, examID uuid.UUID) ([]UngradedEssayRow, error) {
	if _, err := u.exams.FindByID(ctx, examID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	rows, err := u.grading.ListUngradedEssays(ctx, examID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return rows, nil
}

type GradeEssayInput struct {
	QuestionID uuid.UUID
	Score      float64
	Feedback   *string
}

// GradeEssay menyimpan nilai manual esai lalu menjumlahkan ulang skor attempt.
func (u *GradingUsecase) GradeEssay(ctx context.Context, attemptID, questionID uuid.UUID, in GradeEssayInput) (*model.ExamAnswer, error) {
	answer, err := u.grading.FindAnswerByIDByAttempt(ctx, attemptID, questionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Jawaban esai tidak ditemukan pada attempt ini", err)
		}
		return nil, apperror.Internal(err)
	}

	q, err := u.grading.QuestionWithGradingInfo(ctx, questionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Soal tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	if q.QuestionType != model.QuestionTypeEssay {
		return nil, apperror.New(422, "Koreksi manual hanya untuk soal esai", nil)
	}
	if in.Score < 0 || in.Score > q.ScoreWeight {
		return nil, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{
				"score": fmt.Sprintf("nilai harus di antara 0 dan bobot soal (%.2f)", q.ScoreWeight),
			},
		}
	}

	isCorrect := in.Score >= q.ScoreWeight
	if err := u.grading.UpdateGrading(ctx, answer.ID, in.Score, &isCorrect, in.Feedback, "manual"); err != nil {
		return nil, err
	}

	// hitung ulang total skor attempt
	total, err := u.grading.SumScoresByAttempt(ctx, attemptID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if err := u.grading.UpdateAttemptScore(ctx, attemptID, total); err != nil {
		return nil, err
	}

	updated, err := u.grading.FindAnswerByID(ctx, answer.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return updated, nil
}

type SheetOption struct {
	OptionKey string `json:"option_key"`
	Text      string `json:"text"`
}

type SheetAnswer struct {
	QuestionID  uuid.UUID     `json:"question_id"`
	Type        string        `json:"type"`
	Text        string        `json:"text"`
	ScoreWeight float64       `json:"score_weight"`
	AnswerValue string        `json:"answer_value"`
	CorrectKeys []string      `json:"correct_keys"`
	OptionTexts []SheetOption `json:"option_texts"`
	Score       *float64      `json:"score"`
	IsCorrect   *bool         `json:"is_correct"`
	Feedback    *string       `json:"feedback,omitempty"`
	GradedVia   *string       `json:"graded_via,omitempty"`
	AnsweredAt  *time.Time    `json:"answered_at,omitempty"`
}

type SheetStudent struct {
	AttemptID   uuid.UUID     `json:"attempt_id"`
	StudentName string        `json:"student_name"`
	Nis         string        `json:"nis"`
	Status      string        `json:"status"`
	Score       *float64      `json:"score"`
	SubmittedAt *time.Time    `json:"submitted_at"`
	Answers     []SheetAnswer `json:"answers"`
}

// ExamGradingSheet: seluruh jawaban per siswa untuk satu ujian.
func (u *GradingUsecase) ExamGradingSheet(ctx context.Context, examID uuid.UUID) ([]SheetStudent, error) {
	exam, err := u.exams.FindByID(ctx, examID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	attempts, err := u.grading.ListSubmittedAttempts(ctx, exam.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	out := make([]SheetStudent, 0, len(attempts))
	for _, attempt := range attempts {
		answers, err := u.grading.ListAnswersByAttempt(ctx, attempt.ID)
		if err != nil {
			return nil, apperror.Internal(err)
		}

		qIDs := make([]uuid.UUID, 0, len(answers))
		for _, a := range answers {
			qIDs = append(qIDs, a.QuestionID)
		}
		questions, err := u.grading.ListQuestionsByIDs(ctx, qIDs)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		qByID := make(map[uuid.UUID]model.Question, len(questions))
		for _, q := range questions {
			qByID[q.ID] = q
		}

		sheetAnswers := make([]SheetAnswer, 0, len(answers))
		for _, a := range answers {
			q, ok := qByID[a.QuestionID]
			if !ok {
				continue
			}
			item := SheetAnswer{
				QuestionID:  a.QuestionID,
				Type:        strings.ToUpper(q.QuestionType),
				Text:        q.Content,
				ScoreWeight: q.ScoreWeight,
				AnswerValue: a.AnswerValue,
				CorrectKeys: []string{},
				AnsweredAt:  &a.AnsweredAt,
				Score:       a.Score,
				IsCorrect:   a.IsCorrect,
				Feedback:    a.Feedback,
				GradedVia:   a.GradedVia,
			}
			for _, o := range q.Options {
				if o.IsCorrect {
					item.CorrectKeys = append(item.CorrectKeys, strings.ToUpper(o.Label))
				}
				item.OptionTexts = append(item.OptionTexts, SheetOption{
					OptionKey: strings.ToUpper(o.Label),
					Text:      o.Content,
				})
			}
			sort.Slice(item.OptionTexts, func(i, j int) bool {
				return item.OptionTexts[i].OptionKey < item.OptionTexts[j].OptionKey
			})
			sheetAnswers = append(sheetAnswers, item)
		}

		name, nis := "", ""
		if attempt.Student != nil {
			nis = attempt.Student.Nis
			if attempt.Student.User != nil {
				name = attempt.Student.User.Name
			}
		}

		out = append(out, SheetStudent{
			AttemptID:   attempt.ID,
			StudentName: name,
			Nis:         nis,
			Status:      attempt.Status,
			Score:       attempt.Score,
			SubmittedAt: attempt.SubmittedAt,
			Answers:     sheetAnswers,
		})
	}
	return out, nil
}
