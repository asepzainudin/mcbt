package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, m *model.Media) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(m).Error, "")
}

func (r *MediaRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	var m model.Media
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
