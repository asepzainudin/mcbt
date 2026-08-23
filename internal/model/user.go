package model

import (
	"time"
)

type User struct {
	BaseModel
	Username     string     `gorm:"type:varchar(50);not null;uniqueIndex:uq_users_username"`
	Name         string     `gorm:"type:varchar(100);not null"`
	Email        string     `gorm:"type:varchar(150);not null;uniqueIndex:uq_users_email"`
	PasswordHash string     `gorm:"type:varchar(255);not null"`
	IsActive     bool       `gorm:"not null;default:true"`
	TokenVersion int        `gorm:"not null;default:1"`
	LastLoginAt  *time.Time `gorm:"type:timestamptz"`

	Roles   []Role   `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:user_id;joinReferences:role_id"`
	Teacher *Teacher `gorm:"foreignKey:UserID;references:ID"`
	Student *Student `gorm:"foreignKey:UserID;references:ID"`
}

func (User) TableName() string { return "users" }
