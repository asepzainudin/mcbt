package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type SubjectRepository struct {
	db *gorm.DB
}

func NewSubjectRepository(db *gorm.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

func (r *SubjectRepository) List(ctx context.Context, search string, page, limit int) (*PageResult[model.Subject], error) {
	var (
		items []model.Subject
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.Subject{})
	if search != "" {
		q = q.Where("code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	err := q.Order("name ASC").
		Limit(limit).Offset((page - 1) * limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PageResult[model.Subject]{Items: items, TotalItems: total, Page: page, Limit: limit}, nil
}

func (r *SubjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Subject, error) {
	var s model.Subject
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubjectRepository) ExistsDuplicate(ctx context.Context, code string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Subject{}).Where("code = ?", code)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SubjectRepository) Create(ctx context.Context, s *model.Subject) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(s).Error, "Kode mapel sudah digunakan")
}

func (r *SubjectRepository) Update(ctx context.Context, s *model.Subject) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(s).
			Select("code", "name", "description", "updated_at").
			Updates(s).Error,
		"Kode mapel sudah digunakan",
	)
}

func (r *SubjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.Subject{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
