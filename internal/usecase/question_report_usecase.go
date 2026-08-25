package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type QuestionReportUsecase struct {
	reports  *repository.QuestionReportRepository
	attempts *repository.ExamAttemptRepository
	students *repository.StudentRepository
}

func NewQuestionReportUsecase(
	reports *repository.QuestionReportRepository,
	attempts *repository.ExamAttemptRepository,
	students *repository.StudentRepository,
) *QuestionReportUsecase {
	return &QuestionReportUsecase{reports: reports, attempts: attempts, students: students}
}

type CreateReportInput struct {
	AttemptID  uuid.UUID
	QuestionID uuid.UUID
	UserID     uuid.UUID
	Reason     string
}

func (u *QuestionReportUsecase) Create(ctx context.Context, in CreateReportInput) (*model.QuestionReport, error) {
	if strings.TrimSpace(in.Reason) == "" {
		return nil, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"reason": "alasan laporan wajib diisi"},
		}
	}

	// resolve user → student
	student, err := u.students.FindByUserID(ctx, in.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(403, "Akun ini bukan siswa", nil)
		}
		return nil, apperror.Internal(err)
	}

	// cek duplikat
	existing, err := u.reports.FindByAttemptQuestion(ctx, in.AttemptID, in.QuestionID)
	if err == nil {
		return existing, nil // sudah ada → return existing (idempotent)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.Internal(err)
	}

	report := &model.QuestionReport{
		AttemptID:  in.AttemptID,
		QuestionID: in.QuestionID,
		StudentID:  student.ID,
		Reason:     strings.TrimSpace(in.Reason),
		Status:     model.ReportStatusPending,
	}
	if err := u.reports.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

type ReportRow = repository.ReportRow

func (u *QuestionReportUsecase) List(ctx context.Context, status string) ([]ReportRow, error) {
	rows, err := u.reports.ListAll(ctx, status)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return rows, nil
}

type ResolveInput struct {
	Status     string
	Resolution *string
}

func (u *QuestionReportUsecase) Resolve(ctx context.Context, reportID, resolverID uuid.UUID, in ResolveInput) (*model.QuestionReport, error) {
	report, err := u.reports.FindByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Laporan tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	validStatus := map[string]bool{
		model.ReportStatusPending:   true,
		model.ReportStatusReviewing: true,
		model.ReportStatusResolved:  true,
		model.ReportStatusRejected:  true,
	}
	if !validStatus[in.Status] {
		return nil, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"status": "status harus pending, reviewing, resolved, atau rejected"},
		}
	}

	now := time.Now()
	report.Status = in.Status
	report.Resolution = in.Resolution
	report.ResolvedBy = &resolverID
	report.ResolvedAt = &now

	if err := u.reports.Update(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}
