package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

// ImportQuestions persists many questions (with options) in a single transaction.
func (r *QuestionRepository) ImportQuestions(ctx context.Context, questions []model.Question) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for qi := range questions {
			q := &questions[qi]
			if err := tx.Omit("Options").Create(q).Error; err != nil {
				return err
			}
			for i := range q.Options {
				q.Options[i].ID = uuid.New()
				q.Options[i].QuestionID = q.ID
			}
			if len(q.Options) > 0 {
				if err := tx.Create(&q.Options).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}), "")
}
