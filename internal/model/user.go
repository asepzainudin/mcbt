package model

import (
	"time"
)

type User struct {
	BaseModel
	Username     string     `gorm:"type:varchar(50);not null;uniqueIndex:uq_users_username" json:"username"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	Email        string     `gorm:"type:varchar(150);not null;uniqueIndex:uq_users_email" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	TokenVersion int        `gorm:"not null;default:1" json:"-"`
	LastLoginAt  *time.Time `gorm:"type:timestamptz" json:"last_login_at"`

	Roles   []Role   `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:user_id;joinReferences:role_id" json:"roles,omitempty"`
	Teacher *Teacher `gorm:"foreignKey:UserID;references:ID" json:"teacher,omitempty"`
	Student *Student `gorm:"foreignKey:UserID;references:ID" json:"student,omitempty"`
}

func (User) TableName() string { return "users" }
