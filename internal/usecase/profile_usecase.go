package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type ProfileUsecase struct {
	profiles ProfileRepo
	users    UserRepo
}

func NewProfileUsecase(profiles ProfileRepo, users UserRepo) *ProfileUsecase {
	return &ProfileUsecase{profiles: profiles, users: users}
}

func (u *ProfileUsecase) Get(ctx context.Context, userID uuid.UUID) (*repository.ProfileRow, error) {
	if _, err := u.users.FindByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Pengguna tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	row, err := u.profiles.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return row, nil
}

type UpdateProfileInput struct {
	Name  string
	Phone *string
}

func (u *ProfileUsecase) Update(ctx context.Context, userID uuid.UUID, in UpdateProfileInput) (*repository.ProfileRow, error) {
	if _, err := u.users.FindByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Pengguna tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}

	if err := u.profiles.UpdateName(ctx, userID, in.Name); err != nil {
		return nil, apperror.Internal(err)
	}

	// perbarui telepon pada profil siswa/guru bila ada
	if err := u.profiles.UpdateStudentPhone(ctx, userID, in.Phone); err != nil {
		return nil, apperror.Internal(err)
	}
	if err := u.profiles.UpdateTeacherPhone(ctx, userID, in.Phone); err != nil {
		return nil, apperror.Internal(err)
	}

	row, err := u.profiles.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return row, nil
}
