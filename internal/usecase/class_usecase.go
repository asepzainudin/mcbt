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

type ClassUsecase struct {
	repo ClassRepo
	ays  AcademicYearRepo
}

func NewClassUsecase(repo ClassRepo, ays AcademicYearRepo) *ClassUsecase {
	return &ClassUsecase{repo: repo, ays: ays}
}

type ClassInput struct {
	Name           string
	AcademicYearID uuid.UUID
}

func (u *ClassUsecase) List(ctx context.Context, search string, academicYearID *uuid.UUID, page, limit int) ([]model.Class, int64, error) {
	result, err := u.repo.List(ctx, repository.ClassListParams{
		Search:         search,
		AcademicYearID: academicYearID,
		Page:           page,
		Limit:          limit,
	})
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return result.Items, result.TotalItems, nil
}

func (u *ClassUsecase) Create(ctx context.Context, in ClassInput) (*model.Class, error) {
	if _, err := u.ays.FindByID(ctx, in.AcademicYearID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(422, "Tahun ajaran tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	c := &model.Class{Name: in.Name, AcademicYearID: in.AcademicYearID}
	if err := u.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (u *ClassUsecase) Update(ctx context.Context, id uuid.UUID, in ClassInput) (*model.Class, error) {
	c, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Kelas tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	if _, err := u.ays.FindByID(ctx, in.AcademicYearID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(422, "Tahun ajaran tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	c.Name = in.Name
	c.AcademicYearID = in.AcademicYearID
	if err := u.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (u *ClassUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	err := u.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Kelas tidak ditemukan", err)
		}
		return err
	}
	return nil
}
