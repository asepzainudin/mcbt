package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type GradingRepository struct {
	db *gorm.DB
}

func NewGradingRepository(db *gorm.DB) *GradingRepository {
	return &GradingRepository{db: db}
}

// ListSubmittedAttempts returns all submitted attempts of an exam.
func (r *GradingRepository) ListSubmittedAttempts(ctx context.Context, examID uuid.UUID) ([]model.ExamAttempt, error) {
	var attempts []model.ExamAttempt
	err := r.db.WithContext(ctx).
		Preload("Student").
		Where("exam_id = ? AND status = ?", examID, model.AttemptStatusSubmitted).
		Order("submitted_at ASC").
		Find(&attempts).Error
	return attempts, err
}

// ListAnswersByAttempt returns all answers of an attempt.
func (r *GradingRepository) ListAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]model.ExamAnswer, error) {
	var answers []model.ExamAnswer
	err := r.db.WithContext(ctx).
		Where("attempt_id = ?", attemptID).
		Find(&answers).Error
	return answers, err
}

// QuestionWithGradingInfo loads a question + its correct options for grading.
func (r *GradingRepository) QuestionWithGradingInfo(ctx context.Context, questionID uuid.UUID) (*model.Question, error) {
	var q model.Question
	err := r.db.WithContext(ctx).
		Preload("Options", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
		First(&q, "id = ?", questionID).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// UpdateGrading persists grading result of one answer.
func (r *GradingRepository) UpdateGrading(ctx context.Context, answerID uuid.UUID, score float64, isCorrect *bool, feedback *string, via string) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(&model.ExamAnswer{}).
			Where("id = ?", answerID).
			Updates(map[string]any{
				"score":      score,
				"is_correct": isCorrect,
				"feedback":   feedback,
				"graded_at":  time.Now(),
				"graded_via": via,
				"updated_at": time.Now(),
			}).Error,
		"")
}

// UngradedEssayRow is one essay answer awaiting manual grading.
type UngradedEssayRow struct {
	AttemptID    uuid.UUID  `json:"attempt_id"`
	AnswerID     uuid.UUID  `json:"answer_id"`
	StudentName  string     `json:"student_name"`
	Nis          string     `json:"nis"`
	QuestionID   uuid.UUID  `json:"question_id"`
	QuestionText string     `json:"question_text"`
	ScoreWeight  float64    `json:"score_weight"`
	AnswerValue  string     `json:"answer_value"`
	SubmittedAt  *time.Time `json:"submitted_at"`
}

// ListUngradedEssays returns essay answers (submitted attempts) without a grade.
func (r *GradingRepository) ListUngradedEssays(ctx context.Context, examID uuid.UUID) ([]UngradedEssayRow, error) {
	rows := make([]UngradedEssayRow, 0)
	err := r.db.WithContext(ctx).
		Table("exam_answers ea").
		Select(`
			a.id            AS attempt_id,
			ea.id           AS answer_id,
			u.name          AS student_name,
			s.nis           AS nis,
			q.id            AS question_id,
			q.content       AS question_text,
			q.score_weight  AS score_weight,
			ea.answer_value AS answer_value,
			a.submitted_at  AS submitted_at
		`).
		Joins("JOIN exam_attempts a ON a.id = ea.attempt_id").
		Joins("JOIN questions q ON q.id = ea.question_id").
		Joins("JOIN students s ON s.id = a.student_id").
		Joins("JOIN users u ON u.id = s.user_id").
		Where(`
			a.exam_id = ? AND a.status = ?
			AND q.question_type = 'essay'
			AND ea.graded_at IS NULL
		`, examID, model.AttemptStatusSubmitted).
		Order("u.name ASC").
		Scan(&rows).Error
	return rows, err
}

// SumScoresByAttempt returns the sum of graded question scores.
func (r *GradingRepository) SumScoresByAttempt(ctx context.Context, attemptID uuid.UUID) (float64, error) {
	var sum *float64
	err := r.db.WithContext(ctx).Model(&model.ExamAnswer{}).
		Where("attempt_id = ? AND score IS NOT NULL", attemptID).
		Select("COALESCE(SUM(score), 0)").Scan(&sum).Error
	if err != nil {
		return 0, err
	}
	if sum == nil {
		return 0, nil
	}
	return *sum, nil
}

// UpdateAttemptScore writes the final score of an attempt.
func (r *GradingRepository) UpdateAttemptScore(ctx context.Context, attemptID uuid.UUID, score float64) error {
	return r.db.WithContext(ctx).Model(&model.ExamAttempt{}).
		Where("id = ?", attemptID).
		Updates(map[string]any{"score": score, "updated_at": time.Now()}).Error
}

// FindAnswerByID loads one answer.
func (r *GradingRepository) FindAnswerByID(ctx context.Context, id uuid.UUID) (*model.ExamAnswer, error) {
	var a model.ExamAnswer
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

var _ = errors.Is

// FindAnswerByIDByAttempt loads an answer scoped to an attempt + question.
func (r *GradingRepository) FindAnswerByIDByAttempt(ctx context.Context, attemptID, questionID uuid.UUID) (*model.ExamAnswer, error) {
	var a model.ExamAnswer
	err := r.db.WithContext(ctx).
		Where("attempt_id = ? AND question_id = ?", attemptID, questionID).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}
