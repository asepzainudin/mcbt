package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type ExamAttemptRepository struct {
	db *gorm.DB
}

func NewExamAttemptRepository(db *gorm.DB) *ExamAttemptRepository {
	return &ExamAttemptRepository{db: db}
}

func (r *ExamAttemptRepository) FindActive(ctx context.Context, examID, studentID uuid.UUID) (*model.ExamAttempt, error) {
	var a model.ExamAttempt
	err := r.db.WithContext(ctx).
		Where("exam_id = ? AND student_id = ? AND status = ? AND expires_at > ?",
			examID, studentID, model.AttemptStatusInProgress, time.Now()).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ExamAttemptRepository) CountByStudentExam(ctx context.Context, examID, studentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ExamAttempt{}).
		Where("exam_id = ? AND student_id = ?", examID, studentID).
		Count(&count).Error
	return count, err
}

func (r *ExamAttemptRepository) Create(ctx context.Context, a *model.ExamAttempt) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(a).Error, "")
}

// ListActiveByStudent returns published exams (with schedule) where the
// student is a participant, joined with the student's attempt count and any
// active attempt id.
type CandidateExamRow struct {
	ExamID        uuid.UUID  `json:"exam_id"`
	Title         string     `json:"title"`
	SubjectCode   string     `json:"subject_code"`
	SubjectName   string     `json:"subject_name"`
	Duration      int        `json:"duration_minutes"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	TokenEnabled  bool       `json:"token_enabled"`
	MaxAttempts   int        `json:"max_attempts"`
	PassingGrade  float64    `json:"passing_grade"`
	AttemptsUsed  int64      `json:"attempts_used"`
	ActiveAttempt *uuid.UUID `json:"active_attempt_id"`
	ActiveExpires *time.Time `json:"active_expires_at"`
}

func (r *ExamAttemptRepository) ListCandidateExams(ctx context.Context, studentID uuid.UUID) ([]CandidateExamRow, error) {
	rows := make([]CandidateExamRow, 0)
	err := r.db.WithContext(ctx).
		Table("exam_participants ep").
		Select(`
			e.id AS exam_id,
			e.title AS title,
			sub.code AS subject_code,
			sub.name AS subject_name,
			e.duration_minutes AS duration,
			es.start_time AS start_time,
			es.end_time AS end_time,
			e.token_enabled AS token_enabled,
			e.max_attempts AS max_attempts,
			e.passing_grade AS passing_grade,
			(SELECT count(*) FROM exam_attempts ea
			  WHERE ea.exam_id = e.id AND ea.student_id = ep.student_id) AS attempts_used,
			(SELECT ea2.id FROM exam_attempts ea2
			  WHERE ea2.exam_id = e.id AND ea2.student_id = ep.student_id
			    AND ea2.status = 'in_progress' AND ea2.expires_at > now()
			  LIMIT 1) AS active_attempt_id,
			(SELECT ea3.expires_at FROM exam_attempts ea3
			  WHERE ea3.exam_id = e.id AND ea3.student_id = ep.student_id
			    AND ea3.status = 'in_progress' AND ea3.expires_at > now()
			  LIMIT 1) AS active_expires_at
		`).
		Joins("JOIN exams e ON e.id = ep.exam_id").
		Joins("JOIN exam_schedules es ON es.exam_id = e.id").
		Joins("JOIN subjects sub ON sub.id = e.subject_id").
		Where("ep.student_id = ? AND e.status = ?", studentID, model.ExamStatusPublished).
		Order("es.start_time ASC").
		Scan(&rows).Error
	return rows, err
}

var _ = errors.Is

func (r *ExamAttemptRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ExamAttempt, error) {
	var a model.ExamAttempt
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ExamAttemptRepository) MarkExpired(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.ExamAttempt{}).
		Where("id = ? AND status = ?", id, model.AttemptStatusInProgress).
		Update("status", model.AttemptStatusExpired).Error
}

// FinalizeSubmit sets status submitted + submitted_at.
func (r *ExamAttemptRepository) FinalizeSubmit(ctx context.Context, id uuid.UUID, submittedAt time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.ExamAttempt{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       model.AttemptStatusSubmitted,
			"submitted_at": submittedAt,
			"updated_at":   submittedAt,
		})
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
