package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

type QuestionListParams struct {
	BankID      *uuid.UUID
	OwnerUserID *uuid.UUID // data scope guru via bank soal
	Search      string
	Page        int
	Limit       int
	Type        string
}

func (r *QuestionRepository) List(ctx context.Context, p QuestionListParams) (*PageResult[model.Question], error) {
	var (
		items []model.Question
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.Question{})
	if p.BankID != nil {
		q = q.Where("question_bank_id = ?", *p.BankID)
	}
	if p.OwnerUserID != nil {
		q = q.Joins("JOIN question_banks qb ON qb.id = questions.question_bank_id").
			Where("qb.created_by = ?", *p.OwnerUserID)
	}
	if p.Type != "" {
		q = q.Where("questions.question_type = ?", p.Type)
	}
	if p.Search != "" {
		q = q.Where("questions.content ILIKE ?", "%"+p.Search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	err := q.
		Preload("Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Options.Media").
		Preload("Media").
		Order("created_at ASC").
		Limit(p.Limit).Offset((p.Page - 1) * p.Limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PageResult[model.Question]{Items: items, TotalItems: total, Page: p.Page, Limit: p.Limit}, nil
}

func (r *QuestionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Question, error) {
	var q model.Question
	err := r.db.WithContext(ctx).
		Preload("Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Options.Media").
		Preload("Media").
		Preload("QuestionBank").
		First(&q, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuestionRepository) CreateWithOptions(ctx context.Context, question *model.Question) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Options").Create(question).Error; err != nil {
			return err
		}
		for i := range question.Options {
			if question.Options[i].ID == uuid.Nil {
				question.Options[i].ID = uuid.New()
			}
			question.Options[i].QuestionID = question.ID
		}
		if len(question.Options) > 0 {
			if err := tx.Omit("Media").Create(&question.Options).Error; err != nil {
				return err
			}
		}
		return nil
	}), "")
}

// ReplaceOptions deletes existing options and inserts the new set in one transaction.
func (r *QuestionRepository) ReplaceOptions(ctx context.Context, questionID uuid.UUID, options []model.Option) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id = ?", questionID).Delete(&model.Option{}).Error; err != nil {
			return err
		}
		for i := range options {
			options[i].ID = uuid.New()
			options[i].QuestionID = questionID
		}
		if len(options) > 0 {
			if err := tx.Create(&options).Error; err != nil {
				return err
			}
		}
		return nil
	}), "")
}

// UpdateWithOptions updates question meta and replaces all options atomically.
func (r *QuestionRepository) UpdateWithOptions(ctx context.Context, question *model.Question, options []model.Option) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(question).
			Select("question_bank_id", "media_id", "question_type", "content", "score_weight", "explanation", "answer_keys", "media_position", "updated_at").
			Updates(question).Error; err != nil {
			return err
		}
		if err := tx.Where("question_id = ?", question.ID).Delete(&model.Option{}).Error; err != nil {
			return err
		}
		for i := range options {
			options[i].ID = uuid.New()
			options[i].QuestionID = question.ID
		}
		if len(options) > 0 {
			if err := tx.Create(&options).Error; err != nil {
				return err
			}
		}
		return nil
	}), "")
}

func (r *QuestionRepository) ReorderOptions(ctx context.Context, questionID uuid.UUID, orderedIDs []uuid.UUID) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Option{}).Where("question_id = ?", questionID).Count(&count).Error; err != nil {
			return err
		}
		if int(count) != len(orderedIDs) {
			return apperror.Unprocessable(
				"option_ids_order harus memuat semua opsi soal ini tanpa duplikasi", nil)
		}

		if err := tx.Model(&model.Option{}).
			Where("question_id = ?", questionID).
			Update("label", gorm.Expr("(position::int + 1)::text")).Error; err != nil {
			return err
		}

		for i, optID := range orderedIDs {
			res := tx.Model(&model.Option{}).
				Where("id = ? AND question_id = ?", optID, questionID).
				Updates(map[string]any{
					"label":    string(rune('A' + i)),
					"position": i,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	}), "")
}

func (r *QuestionRepository) SetCorrectOption(ctx context.Context, questionID, optionID uuid.UUID) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Option{}).Where("question_id = ?", questionID).Update("is_correct", false).Error; err != nil {
			return err
		}
		res := tx.Model(&model.Option{}).
			Where("id = ? AND question_id = ?", optionID, questionID).
			Update("is_correct", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}), "")
}

func (r *QuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.Question{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *QuestionRepository) CountByBank(ctx context.Context, bankID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Question{}).
		Where("question_bank_id = ?", bankID).
		Count(&count).Error
	return count, err
}
