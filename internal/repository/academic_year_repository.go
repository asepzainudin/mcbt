package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type AcademicYearRepository struct {
	db *gorm.DB
}

func NewAcademicYearRepository(db *gorm.DB) *AcademicYearRepository {
	return &AcademicYearRepository{db: db}
}

func (r *AcademicYearRepository) List(ctx context.Context, search string, page, limit int) (*PageResult[model.AcademicYear], error) {
	var (
		items []model.AcademicYear
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.AcademicYear{})
	if search != "" {
		q = q.Where("year ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	err := q.Order("year DESC").Order("semester ASC").
		Limit(limit).Offset((page - 1) * limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PageResult[model.AcademicYear]{Items: items, TotalItems: total, Page: page, Limit: limit}, nil
}

func (r *AcademicYearRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AcademicYear, error) {
	var ay model.AcademicYear
	err := r.db.WithContext(ctx).First(&ay, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &ay, nil
}

func (r *AcademicYearRepository) ExistsDuplicate(ctx context.Context, year, semester string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.AcademicYear{}).
		Where("year = ? AND semester = ?", year, semester)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AcademicYearRepository) Create(ctx context.Context, ay *model.AcademicYear) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(ay).Error, "Tahun ajaran dengan semester tersebut sudah ada")
}

func (r *AcademicYearRepository) Update(ctx context.Context, ay *model.AcademicYear) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(ay).
			Select("year", "semester", "start_date", "end_date", "updated_at").
			Updates(ay).Error,
		"Tahun ajaran dengan semester tersebut sudah ada",
	)
}

func (r *AcademicYearRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AcademicYear{}).Where("is_active = true").Update("is_active", false).Error; err != nil {
			return err
		}
		res := tx.Model(&model.AcademicYear{}).Where("id = ?", id).Update("is_active", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}), "")
}

func (r *AcademicYearRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.AcademicYear{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
