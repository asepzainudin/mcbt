package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

type SubjectUsecase struct {
	repo SubjectRepo
}

func NewSubjectUsecase(repo SubjectRepo) *SubjectUsecase {
	return &SubjectUsecase{repo: repo}
}

type SubjectInput struct {
	Code        string
	Name        string
	Description *string
}

func (u *SubjectUsecase) List(ctx context.Context, search string, page, limit int) ([]model.Subject, int64, error) {
	result, err := u.repo.List(ctx, search, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return result.Items, result.TotalItems, nil
}

func (u *SubjectUsecase) Create(ctx context.Context, in SubjectInput) (*model.Subject, error) {
	dup, err := u.repo.ExistsDuplicate(ctx, in.Code, nil)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if dup {
		return nil, apperror.New(409, "Kode mapel sudah digunakan", nil)
	}

	s := &model.Subject{Code: in.Code, Name: in.Name, Description: in.Description}
	if err := u.repo.Create(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (u *SubjectUsecase) Update(ctx context.Context, id uuid.UUID, in SubjectInput) (*model.Subject, error) {
	s, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Mapel tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	dup, err := u.repo.ExistsDuplicate(ctx, in.Code, &id)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if dup {
		return nil, apperror.New(409, "Kode mapel sudah digunakan", nil)
	}

	s.Code = in.Code
	s.Name = in.Name
	s.Description = in.Description
	if err := u.repo.Update(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (u *SubjectUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	err := u.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Mapel tidak ditemukan", err)
		}
		return err
	}
	return nil
}
