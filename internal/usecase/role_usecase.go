package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

type RoleUsecase struct {
	roles RoleRepo
	users UserRepo
}

func NewRoleUsecase(roles RoleRepo, users UserRepo) *RoleUsecase {
	return &RoleUsecase{roles: roles, users: users}
}

func (u *RoleUsecase) List(ctx context.Context, page, limit int) ([]model.Role, int64, error) {
	result, err := u.roles.ListPaged(ctx, page, limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}
	return result.Items, result.TotalItems, nil
}

func (u *RoleUsecase) AssignToUser(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	exists, err := u.users.ExistsByID(ctx, userID)
	if err != nil {
		return apperror.Internal(err)
	}
	if !exists {
		return apperror.NotFound("User not found", nil)
	}

	found, err := u.roles.FindByIDs(ctx, roleIDs)
	if err != nil {
		return apperror.Internal(err)
	}
	if len(found) != len(roleIDs) {
		return apperror.New(422, "Some roles do not exist", nil)
	}

	if err := u.roles.ReplaceUserRoles(ctx, userID, roleIDs); err != nil {
		return apperror.Internal(err)
	}
	return nil
}
