package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type ExamAnswerRepository struct {
	db *gorm.DB
}

func NewExamAnswerRepository(db *gorm.DB) *ExamAnswerRepository {
	return &ExamAnswerRepository{db: db}
}

func (r *ExamAnswerRepository) ListByAttempt(ctx context.Context, attemptID uuid.UUID) ([]model.ExamAnswer, error) {
	var answers []model.ExamAnswer
	err := r.db.WithContext(ctx).
		Where("attempt_id = ?", attemptID).
		Find(&answers).Error
	return answers, err
}

func (r *ExamAnswerRepository) FindByAttemptQuestion(ctx context.Context, attemptID, questionID uuid.UUID) (*model.ExamAnswer, error) {
	var a model.ExamAnswer
	err := r.db.WithContext(ctx).
		Where("attempt_id = ? AND question_id = ?", attemptID, questionID).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// UpsertAnswer creates the answer row if missing, then updates its value.
func (r *ExamAnswerRepository) UpsertAnswer(ctx context.Context, attemptID, questionID uuid.UUID, answerValue string, clientTimestamp int64) (*model.ExamAnswer, error) {
	var answer model.ExamAnswer
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where(model.ExamAnswer{AttemptID: attemptID, QuestionID: questionID}).
			FirstOrCreate(&answer)
		if res.Error != nil {
			return res.Error
		}
		answer.AnswerValue = answerValue
		answer.ClientTimestamp = clientTimestamp
		answer.AnsweredAt = time.Now()
		return tx.Model(&answer).
			Select("answer_value", "client_timestamp", "answered_at", "updated_at").
			Updates(&answer).Error
	})
	if err != nil {
		return nil, TranslateDBError(err, "")
	}
	return &answer, nil
}

func (r *ExamAnswerRepository) SetFlag(ctx context.Context, attemptID, questionID uuid.UUID, flagged bool) (*model.ExamAnswer, error) {
	var answer model.ExamAnswer
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where(model.ExamAnswer{AttemptID: attemptID, QuestionID: questionID}).
			FirstOrCreate(&answer)
		if res.Error != nil {
			return res.Error
		}
		answer.IsFlagged = flagged
		return tx.Model(&answer).Update("is_flagged", flagged).Error
	})
	if err != nil {
		return nil, TranslateDBError(err, "")
	}
	return &answer, nil
}

var _ = errors.Is
