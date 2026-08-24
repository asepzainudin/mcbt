package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type QuestionBankRepository struct {
	db *gorm.DB
}

func NewQuestionBankRepository(db *gorm.DB) *QuestionBankRepository {
	return &QuestionBankRepository{db: db}
}

type BankListParams struct {
	Search    string
	SubjectID *uuid.UUID
	Page      int
	Limit     int
}

func (r *QuestionBankRepository) List(ctx context.Context, p BankListParams) (*PageResult[model.QuestionBank], error) {
	var (
		items []model.QuestionBank
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.QuestionBank{})
	if p.SubjectID != nil {
		q = q.Where("subject_id = ?", *p.SubjectID)
	}
	if p.Search != "" {
		q = q.Where("title ILIKE ? OR code ILIKE ?", "%"+p.Search+"%", "%"+p.Search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	err := q.Preload("Subject").Preload("AcademicYear").
		Order("created_at DESC").
		Limit(p.Limit).Offset((p.Page - 1) * p.Limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PageResult[model.QuestionBank]{Items: items, TotalItems: total, Page: p.Page, Limit: p.Limit}, nil
}

func (r *QuestionBankRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.QuestionBank, error) {
	var qb model.QuestionBank
	err := r.db.WithContext(ctx).Preload("Subject").Preload("AcademicYear").First(&qb, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &qb, nil
}

func (r *QuestionBankRepository) Create(ctx context.Context, qb *model.QuestionBank) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(qb).Error, "Code atau judul bank sudah digunakan untuk mapel tersebut")
}

func (r *QuestionBankRepository) Update(ctx context.Context, qb *model.QuestionBank) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(qb).
			Select("subject_id", "academic_year_id", "title", "description", "updated_at").
			Updates(qb).Error,
		"Judul bank sudah digunakan untuk mapel tersebut",
	)
}

func (r *QuestionBankRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.QuestionBank{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *QuestionBankRepository) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).Model(&model.QuestionBank{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()})
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CloneWithQuestions copies the bank (status reset to draft, code suffixed)
// together with every question and its options.
func (r *QuestionBankRepository) CloneWithQuestions(ctx context.Context, source *model.QuestionBank) (*model.QuestionBank, error) {
	var cloneID uuid.UUID
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		clone := model.QuestionBank{
			SubjectID:      source.SubjectID,
			AcademicYearID: source.AcademicYearID,
			Code:           source.Code + "-COPY-" + source.ID.String()[:4],
			Status:         model.BankStatusDraft,
			Title:          source.Title + " (Salinan)",
			Description:    source.Description,
		}
		if err := tx.Create(&clone).Error; err != nil {
			return err
		}
		cloneID = clone.ID

		var questions []model.Question
		if err := tx.Preload("Options").Where("question_bank_id = ?", source.ID).Find(&questions).Error; err != nil {
			return err
		}

		for _, q := range questions {
			newQ := model.Question{
				QuestionBankID: clone.ID,
				MediaID:        q.MediaID,
				QuestionType:   q.QuestionType,
				Content:        q.Content,
				ScoreWeight:    q.ScoreWeight,
				Explanation:    q.Explanation,
				AnswerKeys:     q.AnswerKeys,
			}
			if err := tx.Omit("Options").Create(&newQ).Error; err != nil {
				return err
			}
			for _, o := range q.Options {
				opt := o
				opt.ID = uuid.New()
				opt.QuestionID = newQ.ID
				if err := tx.Omit("Media").Create(&opt).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, TranslateDBError(err, "Gagal mengkloning bank soal")
	}
	return r.FindByID(ctx, cloneID)
}
