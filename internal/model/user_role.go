package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null;primaryKey;index:idx_user_roles_role_id"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
	Role *Role `gorm:"foreignKey:RoleID;references:ID"`
}

func (UserRole) TableName() string { return "user_roles" }
