package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

// AccessUsecase memastikan guru hanya dapat mengelola data
// yang terhubung dengan akunnya sendiri (data scope).
type AccessUsecase struct {
	banks     *repository.QuestionBankRepository
	exams     *repository.ExamRepository
	sections  *repository.ExamSectionRepository
	questions *repository.QuestionRepository
	attempts  *repository.ExamAttemptRepository
}

func NewAccessUsecase(
	banks *repository.QuestionBankRepository,
	exams *repository.ExamRepository,
	sections *repository.ExamSectionRepository,
	questions *repository.QuestionRepository,
	attempts *repository.ExamAttemptRepository,
) *AccessUsecase {
	return &AccessUsecase{banks: banks, exams: exams, sections: sections, questions: questions, attempts: attempts}
}

func errNotOwner() error {
	return apperror.Forbidden("Anda hanya dapat mengelola data milik akun Anda sendiri", nil)
}

func checkBankOwner(creator *uuid.UUID, userID uuid.UUID, isAdmin bool) error {
	if isAdmin {
		return nil
	}
	if creator != nil && *creator == userID {
		return nil
	}
	return errNotOwner()
}

func (u *AccessUsecase) AssertBankOwner(ctx context.Context, userID uuid.UUID, isAdmin bool, bankID uuid.UUID) error {
	bank, err := u.banks.FindByID(ctx, bankID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Bank soal tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return checkBankOwner(bank.CreatedBy, userID, isAdmin)
}

func (u *AccessUsecase) AssertQuestionOwner(ctx context.Context, userID uuid.UUID, isAdmin bool, questionID uuid.UUID) error {
	q, err := u.questions.FindByID(ctx, questionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Soal tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return u.AssertBankOwner(ctx, userID, isAdmin, q.QuestionBankID)
}

// AssertExamOwner: ujian dimiliki oleh pembuatnya.
func (u *AccessUsecase) AssertExamOwner(ctx context.Context, userID uuid.UUID, isAdmin bool, examID uuid.UUID) error {
	exam, err := u.exams.FindByID(ctx, examID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Ujian tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return checkBankOwner(exam.CreatedBy, userID, isAdmin)
}

func (u *AccessUsecase) AssertSectionOwner(ctx context.Context, userID uuid.UUID, isAdmin bool, sectionID uuid.UUID) error {
	section, err := u.sections.FindByID(ctx, sectionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Seksi ujian tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return u.AssertExamOwner(ctx, userID, isAdmin, section.ExamID)
}

func (u *AccessUsecase) AssertAttemptOwner(ctx context.Context, userID uuid.UUID, isAdmin bool, attemptID uuid.UUID) error {
	attempt, err := u.attempts.FindByID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Percobaan ujian tidak ditemukan", err)
		}
		return apperror.Internal(err)
	}
	return u.AssertExamOwner(ctx, userID, isAdmin, attempt.ExamID)
}
