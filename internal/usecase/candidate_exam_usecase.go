package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type CandidateExamUsecase struct {
	attempts     *repository.ExamAttemptRepository
	students     *repository.StudentRepository
	schedules    *repository.ExamScheduleRepository
	exams        *repository.ExamRepository
	participants *repository.ExamParticipantRepository
}

func NewCandidateExamUsecase(
	attempts *repository.ExamAttemptRepository,
	students *repository.StudentRepository,
	schedules *repository.ExamScheduleRepository,
	exams *repository.ExamRepository,
	participants *repository.ExamParticipantRepository,
) *CandidateExamUsecase {
	return &CandidateExamUsecase{
		attempts: attempts, students: students, schedules: schedules,
		exams: exams, participants: participants,
	}
}

// resolveStudent memastikan user adalah siswa dan mengembalikan profilenya.
func (u *CandidateExamUsecase) resolveStudent(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
	student, err := u.students.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(403, "Akun ini bukan siswa", nil)
		}
		return nil, apperror.Internal(err)
	}
	return student, nil
}

func (u *CandidateExamUsecase) ListExams(ctx context.Context, userID uuid.UUID) ([]repository.CandidateExamRow, error) {
	student, err := u.resolveStudent(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := u.attempts.ListCandidateExams(ctx, student.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return rows, nil
}

type candidateContext struct {
	Student  *model.Student
	Exam     *model.Exam
	Schedule *model.ExamSchedule
}

func (u *CandidateExamUsecase) loadContext(ctx context.Context, userID, examID uuid.UUID) (*candidateContext, error) {
	student, err := u.resolveStudent(ctx, userID)
	if err != nil {
		return nil, err
	}

	exam, err := u.exams.FindByID(ctx, examID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	if exam.Status != model.ExamStatusPublished {
		return nil, apperror.New(403, "Ujian belum dibuka", nil)
	}

	// peserta wajib terdaftar pada ujian ini
	assigned, err := u.participants.AssignedStudentIDsByExam(ctx, examID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if !assigned[student.ID] {
		return nil, apperror.New(403, "Anda bukan peserta ujian ini", nil)
	}

	return &candidateContext{Student: student, Exam: exam}, nil
}

func (u *CandidateExamUsecase) ValidateToken(ctx context.Context, userID, examID uuid.UUID, token string) error {
	cc, err := u.loadContext(ctx, userID, examID)
	if err != nil {
		return err
	}

	schedule, err := u.schedules.FindByExam(ctx, examID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(422, "Ujian belum memiliki jadwal", nil)
		}
		return apperror.Internal(err)
	}
	cc.Schedule = schedule

	now := time.Now()
	if now.Before(schedule.StartTime) {
		return apperror.New(403, fmt.Sprintf(
			"Ujian belum dimulai (mulai %s)", schedule.StartTime.Format("02 Jan 2006 15:04")), nil)
	}
	if now.After(schedule.EndTime) {
		return apperror.New(403, "Ujian sudah berakhir", nil)
	}

	if cc.Exam.TokenEnabled {
		if !strings.EqualFold(token, schedule.Token) {
			return apperror.New(403, "Token salah", nil)
		}
	}

	// attempt aktif boleh lanjut tanpa cek kuota
	if _, err := u.attempts.FindActive(ctx, examID, cc.Student.ID); err == nil {
		return nil
	}

	used, err := u.attempts.CountByStudentExam(ctx, examID, cc.Student.ID)
	if err != nil {
		return apperror.Internal(err)
	}
	if int(used) >= cc.Exam.MaxAttempts {
		return apperror.New(403, "Kesempatan mengerjakan sudah habis", nil)
	}
	return nil
}

func (u *CandidateExamUsecase) Start(ctx context.Context, userID, examID uuid.UUID, token string) (*model.ExamAttempt, error) {
	cc, err := u.loadContext(ctx, userID, examID)
	if err != nil {
		return nil, err
	}

	schedule, err := u.schedules.FindByExam(ctx, examID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(422, "Ujian belum memiliki jadwal", nil)
		}
		return nil, apperror.Internal(err)
	}
	cc.Schedule = schedule

	now := time.Now()
	if now.Before(schedule.StartTime) {
		return nil, apperror.New(403, "Ujian belum dimulai", nil)
	}
	if now.After(schedule.EndTime) {
		return nil, apperror.New(403, "Ujian sudah berakhir", nil)
	}

	// resume attempt aktif bila ada
	if active, err := u.attempts.FindActive(ctx, examID, cc.Student.ID); err == nil {
		return active, nil
	}

	if cc.Exam.TokenEnabled {
		if !strings.EqualFold(token, schedule.Token) {
			return nil, apperror.New(403, "Token salah", nil)
		}
	}

	used, err := u.attempts.CountByStudentExam(ctx, examID, cc.Student.ID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if int(used) >= cc.Exam.MaxAttempts {
		return nil, apperror.New(403, "Kesempatan mengerjakan sudah habis", nil)
	}

	attempt := &model.ExamAttempt{
		ExamID:    examID,
		StudentID: cc.Student.ID,
		AttemptNo: int(used) + 1,
		Status:    model.AttemptStatusInProgress,
		StartedAt: now,
		ExpiresAt: now.Add(time.Duration(cc.Exam.DurationMinutes) * time.Minute),
	}
	if err := u.attempts.Create(ctx, attempt); err != nil {
		return nil, err
	}
	return attempt, nil
}
