package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type GradingUsecase struct {
	grading  *repository.GradingRepository
	answers  *repository.ExamAnswerRepository
	attempts *repository.ExamAttemptRepository
	exams    *repository.ExamRepository
}

func NewGradingUsecase(
	grading *repository.GradingRepository,
	answers *repository.ExamAnswerRepository,
	attempts *repository.ExamAttemptRepository,
	exams *repository.ExamRepository,
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

		total := 0.0
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
			total += score
			result.QuestionsGraded++
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
