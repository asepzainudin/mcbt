package model

import (
	"time"

	"github.com/google/uuid"
)

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	CreatedAt    time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (RolePermission) TableName() string { return "role_permissions" }
