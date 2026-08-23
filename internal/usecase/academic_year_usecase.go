package usecase

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

var yearPattern = regexp.MustCompile(`^\d{4}/\d{4}$`)

func validateAYInput(in AcademicYearInput) error {
	if !yearPattern.MatchString(in.Year) {
		return &apperror.AppError{
			Code:    apperror.CodeUnprocessable,
			Message: "Validasi gagal",
			Details: map[string]string{"year": "year harus berformat YYYY/YYYY, contoh: 2025/2026"},
		}
	}
	return nil
}

type AcademicYearUsecase struct {
	repo *repository.AcademicYearRepository
}

func NewAcademicYearUsecase(repo *repository.AcademicYearRepository) *AcademicYearUsecase {
	return &AcademicYearUsecase{repo: repo}
}

type AcademicYearInput struct {
	Year     string
	Semester string
}

func (u *AcademicYearUsecase) List(ctx context.Context, search string, page, limit int) ([]model.AcademicYear, int64, error) {
	result, err := u.repo.List(ctx, search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return result.Items, result.TotalItems, nil
}

func (u *AcademicYearUsecase) Create(ctx context.Context, in AcademicYearInput) (*model.AcademicYear, error) {
	if err := validateAYInput(in); err != nil {
		return nil, err
	}

	dup, err := u.repo.ExistsDuplicate(ctx, in.Year, in.Semester, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if dup {
		return nil, apperror.New(409, "Kombinasi tahun dan semester sudah ada", nil)
	}

	ay := &model.AcademicYear{Year: in.Year, Semester: in.Semester}
	if err := u.repo.Create(ctx, ay); err != nil {
		return nil, err
	}
	return ay, nil
}

func (u *AcademicYearUsecase) Update(ctx context.Context, id uuid.UUID, in AcademicYearInput) (*model.AcademicYear, error) {
	if err := validateAYInput(in); err != nil {
		return nil, err
	}

	ay, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Tahun ajaran tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	dup, err := u.repo.ExistsDuplicate(ctx, in.Year, in.Semester, &id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if dup {
		return nil, apperror.New(409, "Kombinasi tahun dan semester sudah ada", nil)
	}

	ay.Year = in.Year
	ay.Semester = in.Semester
	if err := u.repo.Update(ctx, ay); err != nil {
		return nil, err
	}
	return ay, nil
}

func (u *AcademicYearUsecase) Activate(ctx context.Context, id uuid.UUID) (*model.AcademicYear, error) {
	err := u.repo.Activate(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Tahun ajaran tidak ditemukan", err)
		}
		return nil, err
	}
	return u.repo.FindByID(ctx, id)
}

func (u *AcademicYearUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	err := u.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Tahun ajaran tidak ditemukan", err)
		}
		return err
	}
	return nil
}
