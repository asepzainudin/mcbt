package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type DashboardUsecase struct {
	dash     DashboardRepo
	students StudentRepo
	teachers TeacherRepo
}

func NewDashboardUsecase(
	dash DashboardRepo,
	students StudentRepo,
	teachers TeacherRepo,
) *DashboardUsecase {
	return &DashboardUsecase{dash: dash, students: students, teachers: teachers}
}

func (u *DashboardUsecase) Admin(ctx context.Context) (*repository.AdminStats, error) {
	stats, err := u.dash.AdminStats(ctx)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return stats, nil
}

// Teacher: statistik berbasis bank soal milik guru (user login).
func (u *DashboardUsecase) Teacher(ctx context.Context, userID uuid.UUID) (*repository.TeacherStats, error) {
	teacher, err := u.teachers.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Guru tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	stats, err := u.dash.TeacherStats(ctx, teacher.UserID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return stats, nil
}

// Student: ringkasan ujian untuk siswa yang sedang login.
type StudentSummary struct {
	AssignedExams  int64    `json:"assigned_exams"`
	CompletedExams int64    `json:"completed_exams"`
	PassedExams    int64    `json:"passed_exams"`
	AverageScore   *float64 `json:"average_score"`
	BestScore      *float64 `json:"best_score"`
}

func (u *DashboardUsecase) Student(ctx context.Context, userID uuid.UUID) (*StudentSummary, error) {
	student, err := u.students.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Siswa tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	stats, err := u.dash.StudentStats(ctx, student.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	summary := &StudentSummary{
		AssignedExams:  stats.AssignedExams,
		CompletedExams: stats.CompletedExams,
		PassedExams:    stats.PassedExams,
	}
	if stats.AverageScore.Valid {
		v := stats.AverageScore.Float64
		summary.AverageScore = &v
	}
	if stats.BestScore.Valid {
		v := stats.BestScore.Float64
		summary.BestScore = &v
	}
	return summary, nil
}
