package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type QuestionReportRepository struct {
	db *gorm.DB
}

func NewQuestionReportRepository(db *gorm.DB) *QuestionReportRepository {
	return &QuestionReportRepository{db: db}
}

func (r *QuestionReportRepository) Create(ctx context.Context, report *model.QuestionReport) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(report).Error, "")
}

func (r *QuestionReportRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.QuestionReport, error) {
	var report model.QuestionReport
	err := r.db.WithContext(ctx).
		Preload("Question").
		Preload("Student").
		Preload("Student.User").
		First(&report, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *QuestionReportRepository) FindByAttemptQuestion(ctx context.Context, attemptID, questionID uuid.UUID) (*model.QuestionReport, error) {
	var report model.QuestionReport
	err := r.db.WithContext(ctx).
		Where("attempt_id = ? AND question_id = ?", attemptID, questionID).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

type ReportRow struct {
	ID           uuid.UUID  `json:"id"`
	ExamTitle    string     `json:"exam_title"`
	StudentName  string     `json:"student_name"`
	Nis          string     `json:"nis"`
	QuestionText string     `json:"question_text"`
	Reason       string     `json:"reason"`
	Status       string     `json:"status"`
	Resolution   *string    `json:"resolution,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

func (r *QuestionReportRepository) ListAll(ctx context.Context, status string, ownerUserID *uuid.UUID) ([]ReportRow, error) {
	var rows []ReportRow
	q := r.db.WithContext(ctx).
		Table("question_reports qr").
		Select(`
			qr.id          AS id,
			e.title        AS exam_title,
			u.name         AS student_name,
			s.nis          AS nis,
			q.content      AS question_text,
			qr.reason      AS reason,
			qr.status      AS status,
			qr.resolution  AS resolution,
			qr.created_at  AS created_at,
			qr.resolved_at AS resolved_at
		`).
		Joins("JOIN exam_attempts a ON a.id = qr.attempt_id").
		Joins("JOIN exams e ON e.id = a.exam_id").
		Joins("JOIN questions q ON q.id = qr.question_id").
		Joins("JOIN students s ON s.id = qr.student_id").
		Joins("JOIN users u ON u.id = s.user_id").
		Order("qr.created_at DESC")

	if status != "" {
		q = q.Where("qr.status = ?", status)
	}
	if ownerUserID != nil {
		// guru menangani laporan pada ujian buatannya sendiri
		q = q.Where("e.created_by = ?", *ownerUserID)
	}
	err := q.Scan(&rows).Error
	return rows, err
}

func (r *QuestionReportRepository) Update(ctx context.Context, report *model.QuestionReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}
