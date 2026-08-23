package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type ClassRepository struct {
	db *gorm.DB
}

func NewClassRepository(db *gorm.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

type ClassListParams struct {
	Search         string
	AcademicYearID *uuid.UUID
	Page           int
	Limit          int
}

func (r *ClassRepository) List(ctx context.Context, p ClassListParams) (*PageResult[model.Class], error) {
	var (
		items []model.Class
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.Class{})
	if p.AcademicYearID != nil {
		q = q.Where("academic_year_id = ?", *p.AcademicYearID)
	}
	if p.Search != "" {
		q = q.Where("name ILIKE ?", "%"+p.Search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	err := q.Preload("AcademicYear").
		Order("name ASC").
		Limit(p.Limit).Offset((p.Page - 1) * p.Limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PageResult[model.Class]{Items: items, TotalItems: total, Page: p.Page, Limit: p.Limit}, nil
}

func (r *ClassRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Class, error) {
	var c model.Class
	err := r.db.WithContext(ctx).Preload("AcademicYear").First(&c, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ClassRepository) ListAll(ctx context.Context) ([]model.Class, error) {
	var classes []model.Class
	err := r.db.WithContext(ctx).
		Preload("AcademicYear").
		Order("name ASC").
		Find(&classes).Error
	return classes, err
}

func (r *ClassRepository) ExistsDuplicate(ctx context.Context, academicYearID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Class{}).
		Where("academic_year_id = ? AND name = ?", academicYearID, name)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ClassRepository) Create(ctx context.Context, c *model.Class) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(c).Error, "Nama kelas sudah digunakan di tahun ajaran tersebut")
}

func (r *ClassRepository) Update(ctx context.Context, c *model.Class) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(c).
			Select("academic_year_id", "name", "grade_level", "homeroom_teacher_id", "updated_at").
			Updates(c).Error,
		"Nama kelas sudah digunakan di tahun ajaran tersebut",
	)
}

func (r *ClassRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.Class{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
