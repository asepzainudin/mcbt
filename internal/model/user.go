package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	BaseModel
	RoleID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_users_role_id"`
	Name         string     `gorm:"type:varchar(100);not null"`
	Email        string     `gorm:"type:varchar(150);not null;uniqueIndex:uq_users_email"`
	PasswordHash string     `gorm:"type:varchar(255);not null"`
	IsActive     bool       `gorm:"not null;default:true"`
	LastLoginAt  *time.Time `gorm:"type:timestamptz"`

	Role    Role     `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Teacher *Teacher `gorm:"foreignKey:UserID;references:ID"`
	Student *Student `gorm:"foreignKey:UserID;references:ID"`
}

func (User) TableName() string { return "users" }
