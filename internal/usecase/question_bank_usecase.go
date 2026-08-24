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

type QuestionBankUsecase struct {
	repo      *repository.QuestionBankRepository
	subjects  *repository.SubjectRepository
	questions *repository.QuestionRepository
}

func NewQuestionBankUsecase(
	repo *repository.QuestionBankRepository,
	subjects *repository.SubjectRepository,
	questions *repository.QuestionRepository,
) *QuestionBankUsecase {
	return &QuestionBankUsecase{repo: repo, subjects: subjects, questions: questions}
}

type QuestionBankInput struct {
	SubjectID      uuid.UUID
	AcademicYearID *uuid.UUID
	Code           string
	Title          string
	Description    *string
}

func (u *QuestionBankUsecase) List(ctx context.Context, search string, subjectID *uuid.UUID, page, limit int) ([]model.QuestionBank, int64, error) {
	result, err := u.repo.List(ctx, repository.BankListParams{
		Search:    search,
		SubjectID: subjectID,
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return result.Items, result.TotalItems, nil
}

func (u *QuestionBankUsecase) Get(ctx context.Context, id uuid.UUID) (*model.QuestionBank, error) {
	qb, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Bank soal tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	return qb, nil
}

func (u *QuestionBankUsecase) subjectExists(ctx context.Context, id uuid.UUID) error {
	_, err := u.subjects.FindByID(ctx, id)
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"subject_id": "mapel tidak ditemukan"},
		}
	}
	return apperror.Internal(err)
}

func (u *QuestionBankUsecase) Create(ctx context.Context, in QuestionBankInput) (*model.QuestionBank, error) {
	if err := u.subjectExists(ctx, in.SubjectID); err != nil {
		return nil, err
	}

	qb := &model.QuestionBank{
		SubjectID:      in.SubjectID,
		AcademicYearID: in.AcademicYearID,
		Code:           in.Code,
		Status:         model.BankStatusDraft,
		Title:          in.Title,
		Description:    in.Description,
	}
	if err := u.repo.Create(ctx, qb); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, qb.ID)
}

func (u *QuestionBankUsecase) Update(ctx context.Context, id uuid.UUID, in QuestionBankInput) (*model.QuestionBank, error) {
	qb, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Bank soal tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	if err := u.subjectExists(ctx, in.SubjectID); err != nil {
		return nil, err
	}

	qb.SubjectID = in.SubjectID
	qb.AcademicYearID = in.AcademicYearID
	qb.Title = in.Title
	qb.Description = in.Description
	if err := u.repo.Update(ctx, qb); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *QuestionBankUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	err := u.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Bank soal tidak ditemukan", err)
		}
		return err
	}
	return nil
}

func (u *QuestionBankUsecase) Clone(ctx context.Context, id uuid.UUID) (*model.QuestionBank, error) {
	source, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Bank soal tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	clone, err := u.repo.CloneWithQuestions(ctx, source)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

func (u *QuestionBankUsecase) Publish(ctx context.Context, id uuid.UUID) (*model.QuestionBank, error) {
	if _, err := u.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Bank soal tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	count, err := u.questions.CountByBank(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if count == 0 {
		return nil, apperror.New(422, "Bank soal masih kosong — tambahkan minimal 1 soal sebelum publish", nil)
	}

	if err := u.repo.SetStatus(ctx, id, model.BankStatusPublished); err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *QuestionBankUsecase) Archive(ctx context.Context, id uuid.UUID) (*model.QuestionBank, error) {
	if err := u.repo.SetStatus(ctx, id, model.BankStatusArchived); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Bank soal tidak ditemukan", err)
		}
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}
