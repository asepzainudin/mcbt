package usecase

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type ResultUsecase struct {
	results  ResultRepo
	exams    ExamRepo
	students StudentRepo
}

func NewResultUsecase(
	results ResultRepo,
	exams ExamRepo,
	students StudentRepo,
) *ResultUsecase {
	return &ResultUsecase{results: results, exams: exams, students: students}
}

type RankedResult struct {
	Rank         int        `json:"rank"`
	AttemptID    uuid.UUID  `json:"attempt_id"`
	StudentID    uuid.UUID  `json:"student_id"`
	StudentName  string     `json:"student_name"`
	Nis          string     `json:"nis"`
	ClassName    *string    `json:"class_name"`
	Status       string     `json:"status"`
	Score        *float64   `json:"score"`
	PassingGrade float64    `json:"passing_grade"`
	Passed       bool       `json:"passed"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	AttemptsUsed int64      `json:"attempts_used"`
}

func (u *ResultUsecase) ExamResults(ctx context.Context, examID uuid.UUID, classID *uuid.UUID) ([]RankedResult, error) {
	if _, err := u.exams.FindByID(ctx, examID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	rows, err := u.results.ExamResults(ctx, examID, classID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	ranked := make([]RankedResult, 0, len(rows))
	for i, r := range rows {
		passed := r.Score != nil && *r.Score >= r.PassingGrade
		ranked = append(ranked, RankedResult{
			Rank:         i + 1,
			AttemptID:    r.AttemptID,
			StudentID:    r.StudentID,
			StudentName:  r.StudentName,
			Nis:          r.Nis,
			ClassName:    r.ClassName,
			Status:       r.Status,
			Score:        r.Score,
			PassingGrade: r.PassingGrade,
			Passed:       passed,
			SubmittedAt:  r.SubmittedAt,
			AttemptsUsed: r.AttemptsCount,
		})
	}
	_ = sort.Slice
	return ranked, nil
}

// StudentResults returns a student's results across exams.
// Score hanya tampil bila results_published atau show_result_immediately.
func (u *ResultUsecase) StudentResults(ctx context.Context, studentID uuid.UUID) ([]repository.StudentResultRow, error) {
	if _, err := u.students.FindByID(ctx, studentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Siswa tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	rows, err := u.results.StudentResults(ctx, studentID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	// sembunyikan skor bila hasil belum dipublikasikan & tampil langsung tidak aktif
	for i := range rows {
		if !rows[i].ResultsPublished && !rows[i].ShowResultImmediately {
			rows[i].Score = nil
		}
	}
	return rows, nil
}

// PublishResults toggle hasil ujian.
func (u *ResultUsecase) PublishResults(ctx context.Context, examID uuid.UUID, published bool) error {
	if _, err := u.exams.FindByID(ctx, examID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return u.results.SetResultsPublished(ctx, examID, published)
}

// ExamReport returns paginated results across all exams with filters.
func (u *ResultUsecase) ExamReport(ctx context.Context, f repository.ExamReportFilter) (*repository.ExamReportPageResult, error) {
	return u.results.ExamReport(ctx, f)
}
