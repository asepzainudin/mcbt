package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

type AdminStats struct {
	TotalStudents      int64 `json:"total_students"`
	TotalTeachers      int64 `json:"total_teachers"`
	TotalQuestionBanks int64 `json:"total_question_banks"`
	PublishedBanks     int64 `json:"published_banks"`
	TotalExams         int64 `json:"total_exams"`
	PublishedExams     int64 `json:"published_exams"`
	OngoingExams       int64 `json:"ongoing_exams"`
	TotalAttempts      int64 `json:"total_attempts"`
}

func (r *DashboardRepository) AdminStats(ctx context.Context) (*AdminStats, error) {
	s := &AdminStats{}
	err := r.db.WithContext(ctx).
		Raw(`SELECT
			(SELECT COUNT(*) FROM students) AS total_students,
			(SELECT COUNT(*) FROM teachers) AS total_teachers,
			(SELECT COUNT(*) FROM question_banks) AS total_question_banks,
			(SELECT COUNT(*) FROM question_banks WHERE status = 'published') AS published_banks,
			(SELECT COUNT(*) FROM exams) AS total_exams,
			(SELECT COUNT(*) FROM exams WHERE status = 'published') AS published_exams,
			(SELECT COUNT(DISTINCT e.id) FROM exams e
				JOIN exam_schedules es ON es.exam_id = e.id
				WHERE e.status = 'published' AND NOW() BETWEEN es.start_time AND es.end_time) AS ongoing_exams,
			(SELECT COUNT(*) FROM exam_attempts WHERE status = 'submitted') AS total_attempts
		`).Scan(s).Error
	if err != nil {
		return nil, err
	}
	return s, nil
}

type TeacherStats struct {
	TotalBanks     int64 `json:"total_banks"`
	PublishedBanks int64 `json:"published_banks"`
	TotalQuestions int64 `json:"total_questions"`
	TotalExams     int64 `json:"total_exams"`
	PublishedExams int64 `json:"published_exams"`
	TotalStudents  int64 `json:"total_students"`
}

func (r *DashboardRepository) TeacherStats(ctx context.Context, userID uuid.UUID) (*TeacherStats, error) {
	s := &TeacherStats{}
	err := r.db.WithContext(ctx).
		Raw(`SELECT
			(SELECT COUNT(*) FROM question_banks WHERE created_by = ?) AS total_banks,
			(SELECT COUNT(*) FROM question_banks WHERE created_by = ? AND status = 'published') AS published_banks,
			(SELECT COUNT(*) FROM questions q JOIN question_banks b ON b.id = q.question_bank_id WHERE b.created_by = ?) AS total_questions,
			(SELECT COUNT(DISTINCT e.id) FROM exams e JOIN question_banks b ON b.id = e.question_bank_id WHERE b.created_by = ?) AS total_exams,
			(SELECT COUNT(DISTINCT e.id) FROM exams e JOIN question_banks b ON b.id = e.question_bank_id WHERE b.created_by = ? AND e.status = 'published') AS published_exams,
			(SELECT COUNT(*) FROM students) AS total_students
		`, userID, userID, userID, userID, userID).Scan(s).Error
	if err != nil {
		return nil, err
	}
	return s, nil
}

type StudentStats struct {
	AssignedExams  int64           `json:"assigned_exams" gorm:"column:assigned_exams"`
	CompletedExams int64           `json:"completed_exams" gorm:"column:completed_exams"`
	AverageScore   sql.NullFloat64 `gorm:"column:avg_score" json:"-"`
	BestScore      sql.NullFloat64 `gorm:"column:max_score" json:"-"`
	PassedExams    int64           `json:"passed_exams" gorm:"column:passed_exams"`
}

// StudentBestScores: satu baris terbaik per ujian untuk siswa.
func (r *DashboardRepository) StudentStats(ctx context.Context, studentID uuid.UUID) (*StudentStats, error) {
	s := &StudentStats{}
	bestPerExam := `
		SELECT a.exam_id, MAX(a.score) AS best_score, MAX(e.passing_grade) AS passing_grade
		FROM exam_attempts a
		JOIN exams e ON e.id = a.exam_id
		WHERE a.student_id = ? AND a.status = ?
		GROUP BY a.exam_id`
	b := bestPerExam
	args := []any{studentID}
	args = append(args, studentID, model.AttemptStatusSubmitted)
	args = append(args, studentID, model.AttemptStatusSubmitted)
	args = append(args, studentID, model.AttemptStatusSubmitted)
	args = append(args, studentID, model.AttemptStatusSubmitted)

	err := r.db.WithContext(ctx).
		Raw(`SELECT
				COALESCE((SELECT COUNT(DISTINCT ep.exam_id) FROM exam_participants ep
					JOIN exams e ON e.id = ep.exam_id
					WHERE ep.student_id = ? AND e.status = 'published'), 0) AS assigned_exams,
				COALESCE((SELECT COUNT(*) FROM (`+b+`) t), 0) AS completed_exams,
				(SELECT COALESCE(AVG(t.best_score), 0) FROM (`+b+`) t) AS avg_score,
				(SELECT COALESCE(MAX(t.best_score), 0) FROM (`+b+`) t) AS max_score,
				COALESCE((SELECT COUNT(*) FROM (`+b+`) t WHERE t.best_score >= t.passing_grade), 0) AS passed_exams
		`, args...).
		Scan(s).Error
	if err != nil {
		return nil, err
	}
	return s, nil
}
