package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

type PageResult[T any] struct {
	Items      []T
	TotalItems int64
	Page       int
	Limit      int
}

func (r *RoleRepository) ListPaged(ctx context.Context, page, limit int) (*PageResult[model.Role], error) {
	var (
		roles []model.Role
		total int64
	)

	if err := r.db.WithContext(ctx).Model(&model.Role{}).Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Model(&model.Role{}).
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return &PageResult[model.Role]{
		Items:      roles,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (r *RoleRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) ReplaceUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			link := model.UserRole{UserID: userID, RoleID: roleID}
			if err := tx.Create(&link).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
