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
	ExamID                uuid.UUID  `json:"exam_id"`
	Title                 string     `json:"title"`
	SubjectCode           string     `json:"subject_code"`
	SubjectName           string     `json:"subject_name"`
	Duration              int        `json:"duration_minutes"`
	StartTime             time.Time  `json:"start_time"`
	EndTime               time.Time  `json:"end_time"`
	TokenEnabled          bool       `json:"token_enabled"`
	MaxAttempts           int        `json:"max_attempts"`
	PassingGrade          float64    `json:"passing_grade"`
	AttemptsUsed          int64      `json:"attempts_used"`
	ActiveAttempt         *uuid.UUID `json:"active_attempt_id"`
	ActiveExpires         *time.Time `json:"active_expires_at"`
	LastStatus            string     `json:"last_status"`
	Score                 *float64   `json:"score"`
	SubmittedAt           *time.Time `json:"submitted_at"`
	ShowResultImmediately bool       `json:"show_result_immediately"`
	HasEssay              bool       `json:"has_essay"`
	EssayUngraded         bool       `json:"essay_ungraded"`
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
			e.show_result_immediately AS show_result_immediately,
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
			  LIMIT 1) AS active_expires_at,
			(SELECT ea4.status FROM exam_attempts ea4
			  WHERE ea4.exam_id = e.id AND ea4.student_id = ep.student_id
			  ORDER BY ea4.attempt_no DESC LIMIT 1) AS last_status,
			(SELECT ea5.score FROM exam_attempts ea5
			  WHERE ea5.exam_id = e.id AND ea5.student_id = ep.student_id
			  ORDER BY ea5.attempt_no DESC LIMIT 1) AS score,
			(SELECT ea6.submitted_at FROM exam_attempts ea6
			  WHERE ea6.exam_id = e.id AND ea6.student_id = ep.student_id
			  ORDER BY ea6.attempt_no DESC LIMIT 1) AS submitted_at,
			(EXISTS (SELECT 1 FROM questions q
			  WHERE q.question_bank_id = e.question_bank_id AND q.question_type = 'essay')
			  OR EXISTS (SELECT 1 FROM exam_section_questions esq
			  JOIN questions q ON q.id = esq.question_id
			  JOIN exam_sections es ON es.id = esq.section_id
			  WHERE es.exam_id = e.id AND q.question_type = 'essay')) AS has_essay,
			EXISTS (SELECT 1
			  FROM exam_answers ea
			  JOIN exam_attempts a ON a.id = ea.attempt_id
			  JOIN questions q ON q.id = ea.question_id
			  WHERE a.exam_id = e.id AND a.student_id = ep.student_id
			    AND a.status = 'submitted'
			    AND q.question_type = 'essay'
			    AND ea.graded_at IS NULL) AS essay_ungraded
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
