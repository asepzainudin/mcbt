package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type ExamScheduleRepository struct {
	db *gorm.DB
}

func NewExamScheduleRepository(db *gorm.DB) *ExamScheduleRepository {
	return &ExamScheduleRepository{db: db}
}

func (r *ExamScheduleRepository) FindByExam(ctx context.Context, examID uuid.UUID) (*model.ExamSchedule, error) {
	var s model.ExamSchedule
	err := r.db.WithContext(ctx).First(&s, "exam_id = ?", examID).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ExamScheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ExamSchedule, error) {
	var s model.ExamSchedule
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ExamScheduleRepository) Create(ctx context.Context, s *model.ExamSchedule) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(s).Error, "Jadwal untuk ujian ini sudah ada")
}

func (r *ExamScheduleRepository) Update(ctx context.Context, s *model.ExamSchedule) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(s).
			Select("start_time", "end_time", "token", "updated_at").
			Updates(s).Error,
		"Token sudah digunakan jadwal lain",
	)
}

func (r *ExamScheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.ExamSchedule{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ExamScheduleRepository) TokenExists(ctx context.Context, token string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.ExamSchedule{}).Where("token = ?", token)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}
