package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type ExamRepository struct {
	db *gorm.DB
}

func NewExamRepository(db *gorm.DB) *ExamRepository {
	return &ExamRepository{db: db}
}

type ExamListParams struct {
	Search         string
	SubjectID      *uuid.UUID
	AcademicYearID *uuid.UUID
	Status         string
	Page           int
	Limit          int
}

func (r *ExamRepository) List(ctx context.Context, p ExamListParams) (*PageResult[model.Exam], error) {
	var (
		items []model.Exam
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.Exam{})
	if p.SubjectID != nil {
		q = q.Where("subject_id = ?", *p.SubjectID)
	}
	if p.AcademicYearID != nil {
		q = q.Where("academic_year_id = ?", *p.AcademicYearID)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}
	if p.Search != "" {
		q = q.Where("title ILIKE ?", "%"+p.Search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	err := q.Preload("Subject").Preload("AcademicYear").Preload("QuestionBank").
		Order("created_at DESC").
		Limit(p.Limit).Offset((p.Page - 1) * p.Limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PageResult[model.Exam]{Items: items, TotalItems: total, Page: p.Page, Limit: p.Limit}, nil
}

func (r *ExamRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Exam, error) {
	var e model.Exam
	err := r.db.WithContext(ctx).
		Preload("Subject").Preload("AcademicYear").Preload("QuestionBank").
		First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExamRepository) Create(ctx context.Context, e *model.Exam) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(e).Error, "")
}

func (r *ExamRepository) UpdateCore(ctx context.Context, e *model.Exam) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(e).
			Select("title", "description", "subject_id", "academic_year_id", "question_bank_id", "updated_at").
			Updates(e).Error,
		"")
}

func (r *ExamRepository) UpdateSettings(ctx context.Context, e *model.Exam) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(e).
			Select(
				"duration_minutes", "max_attempts", "passing_grade",
				"randomize_questions", "randomize_options", "allow_backtrack",
				"auto_submit", "show_result_immediately",
				"negative_marking", "negative_value",
				"token_enabled", "exam_token", "updated_at",
			).
			Updates(e).Error,
		"")
}

func (r *ExamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.Exam{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
