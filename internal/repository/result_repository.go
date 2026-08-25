package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type ResultRepository struct {
	db *gorm.DB
}

func NewResultRepository(db *gorm.DB) *ResultRepository {
	return &ResultRepository{db: db}
}

type ExamResultRow struct {
	AttemptID     uuid.UUID  `json:"attempt_id"`
	StudentID     uuid.UUID  `json:"student_id"`
	StudentName   string     `json:"student_name"`
	Nis           string     `json:"nis"`
	ClassName     *string    `json:"class_name"`
	Status        string     `json:"status"`
	Score         *float64   `json:"score"`
	PassingGrade  float64    `json:"passing_grade"`
	SubmittedAt   *time.Time `json:"submitted_at"`
	AttemptsCount int64      `json:"attempts_used"`
}

func (r *ResultRepository) ExamResults(ctx context.Context, examID uuid.UUID, classID *uuid.UUID) ([]ExamResultRow, error) {
	var rows []ExamResultRow
	q := r.db.WithContext(ctx).
		Table("exam_attempts a").
		Select(`
			a.id          AS attempt_id,
			a.student_id  AS student_id,
			u.name        AS student_name,
			s.nis         AS nis,
			c.name        AS class_name,
			a.status      AS status,
			a.score       AS score,
			e.passing_grade AS passing_grade,
			a.submitted_at  AS submitted_at,
			(SELECT count(*) FROM exam_attempts ea WHERE ea.exam_id = a.exam_id AND ea.student_id = a.student_id) AS attempts_count
		`).
		Joins("JOIN students s ON s.id = a.student_id").
		Joins("JOIN users u ON u.id = s.user_id").
		Joins("LEFT JOIN classes c ON c.id = s.class_id").
		Joins("JOIN exams e ON e.id = a.exam_id").
		Where("a.exam_id = ? AND a.status = ?", examID, model.AttemptStatusSubmitted).
		Order("a.score DESC NULLS LAST, u.name ASC")

	if classID != nil {
		q = q.Where("s.class_id = ?", *classID)
	}
	err := q.Scan(&rows).Error
	return rows, err
}

// StudentResultRow is one submitted attempt of a student across exams.
type StudentResultRow struct {
	ExamID                uuid.UUID  `json:"exam_id"`
	ExamTitle             string     `json:"exam_title"`
	SubjectName           string     `json:"subject_name"`
	Status                string     `json:"status"`
	Score                 *float64   `json:"score"`
	PassingGrade          float64    `json:"passing_grade"`
	ResultsPublished      bool       `json:"results_published"`
	ShowResultImmediately bool       `json:"show_result_immediately"`
	HasEssay              bool       `json:"has_essay"`
	EssayUngraded         bool       `json:"essay_ungraded"`
	SubmittedAt           *time.Time `json:"submitted_at"`
}

func (r *ResultRepository) StudentResults(ctx context.Context, studentID uuid.UUID) ([]StudentResultRow, error) {
	var rows []StudentResultRow
	err := r.db.WithContext(ctx).
		Table("exam_attempts a").
		Select(`
			e.id             AS exam_id,
			e.title          AS exam_title,
			sub.name         AS subject_name,
			a.status         AS status,
			a.score          AS score,
			e.passing_grade  AS passing_grade,
			e.results_published AS results_published,
			e.show_result_immediately AS show_result_immediately,
			a.submitted_at   AS submitted_at,
			(EXISTS (SELECT 1 FROM questions q
				WHERE q.question_bank_id = e.question_bank_id AND q.question_type = 'essay')
			 OR EXISTS (SELECT 1 FROM exam_section_questions esq
				JOIN questions q ON q.id = esq.question_id
				JOIN exam_sections es ON es.id = esq.section_id
				WHERE es.exam_id = e.id AND q.question_type = 'essay')) AS has_essay,
			EXISTS (SELECT 1
				FROM exam_answers ea
				JOIN questions q ON q.id = ea.question_id
				WHERE ea.attempt_id = a.id AND q.question_type = 'essay' AND ea.graded_at IS NULL) AS essay_ungraded
		`).
		Joins("JOIN exams e ON e.id = a.exam_id").
		Joins("JOIN subjects sub ON sub.id = e.subject_id").
		Where("a.student_id = ? AND a.status = ?", studentID, model.AttemptStatusSubmitted).
		Order("a.submitted_at DESC").
		Scan(&rows).Error
	return rows, err
}

// PublishResults toggles results visibility for an exam.
func (r *ResultRepository) SetResultsPublished(ctx context.Context, examID uuid.UUID, published bool) error {
	return r.db.WithContext(ctx).Model(&model.Exam{}).
		Where("id = ?", examID).
		Updates(map[string]any{"results_published": published, "updated_at": time.Now()}).Error
}

// GetExam loads an exam.
func (r *ResultRepository) GetExam(ctx context.Context, examID uuid.UUID) (*model.Exam, error) {
	var e model.Exam
	err := r.db.WithContext(ctx).First(&e, "id = ?", examID).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}
