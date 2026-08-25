package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type ExamParticipantUsecase struct {
	participants *repository.ExamParticipantRepository
	exams        *repository.ExamRepository
	classes      *repository.ClassRepository
	students     *repository.StudentRepository
}

func NewExamParticipantUsecase(
	participants *repository.ExamParticipantRepository,
	exams *repository.ExamRepository,
	classes *repository.ClassRepository,
	students *repository.StudentRepository,
) *ExamParticipantUsecase {
	return &ExamParticipantUsecase{
		participants: participants,
		exams:        exams,
		classes:      classes,
		students:     students,
	}
}

func (u *ExamParticipantUsecase) examExists(ctx context.Context, examID uuid.UUID) error {
	if _, err := u.exams.FindByID(ctx, examID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return nil
}

type AssignResult struct {
	Assigned int `json:"assigned"`
	Skipped  int `json:"skipped"`
}

// AssignClasses expands class_ids into their students and assigns them all.
func (u *ExamParticipantUsecase) AssignClasses(ctx context.Context, examID uuid.UUID, classIDs []uuid.UUID) (*AssignResult, error) {
	if err := u.examExists(ctx, examID); err != nil {
		return nil, err
	}
	if len(classIDs) == 0 {
		return nil, apperror.New(422, "Pilih minimal satu kelas", nil)
	}
	for _, cid := range classIDs {
		exists, err := u.classes.ExistsByID(ctx, cid)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		if !exists {
			return nil, &apperror.AppError{
				Code:    apperror.CodeUnprocessable,
				Message: "Validasi gagal",
				Details: map[string]string{"class_ids": "kelas tidak ditemukan"},
			}
		}
	}

	studentIDs, err := u.participants.StudentIDsByClasses(ctx, classIDs)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if len(studentIDs) == 0 {
		return &AssignResult{Assigned: 0, Skipped: 0}, nil
	}

	assigned, err := u.participants.Assign(ctx, examID, studentIDs, model.AssignedViaClass)
	if err != nil {
		return nil, err
	}
	return &AssignResult{Assigned: assigned, Skipped: len(studentIDs) - assigned}, nil
}

// AssignIndividuals assigns specific students manually.
func (u *ExamParticipantUsecase) AssignIndividuals(ctx context.Context, examID uuid.UUID, studentIDs []uuid.UUID) (*AssignResult, error) {
	if err := u.examExists(ctx, examID); err != nil {
		return nil, err
	}
	if len(studentIDs) == 0 {
		return nil, apperror.New(422, "Pilih minimal satu siswa", nil)
	}
	exists, err := u.participants.StudentsExist(ctx, studentIDs)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if !exists {
		return nil, &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"student_ids": "ada siswa yang tidak ditemukan"},
		}
	}

	assigned, err := u.participants.Assign(ctx, examID, studentIDs, model.AssignedViaIndividual)
	if err != nil {
		return nil, err
	}
	return &AssignResult{Assigned: assigned, Skipped: len(studentIDs) - assigned}, nil
}

func (u *ExamParticipantUsecase) List(ctx context.Context, examID uuid.UUID) ([]model.ExamParticipant, error) {
	if err := u.examExists(ctx, examID); err != nil {
		return nil, err
	}
	participants, err := u.participants.ListByExam(ctx, examID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return participants, nil
}

// Remove menghapus peserta BESERTA seluruh attempt & jawabannya pada ujian ini,
// sehingga saat di-assign kembali siswa dapat mengerjakan dari awal.
func (u *ExamParticipantUsecase) Remove(ctx context.Context, examID, participantID uuid.UUID) error {
	if err := u.examExists(ctx, examID); err != nil {
		return err
	}

	participant, err := u.participants.FindByID(ctx, participantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Peserta tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	if participant.ExamID != examID {
		return apperror.NotFound("Peserta tidak ditemukan pada ujian ini", nil)
	}

	if err := u.participants.RemoveWithCleanup(ctx, examID, participantID, participant.StudentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Peserta tidak ditemukan", err)
		}
		return err
	}
	return nil
}
