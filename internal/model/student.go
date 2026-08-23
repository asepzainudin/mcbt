package model

import (
	"github.com/google/uuid"
)

type Student struct {
	BaseModel
	UserID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_students_user_id" json:"user_id"`
	ClassID *uuid.UUID `gorm:"type:uuid;index:idx_students_class_id" json:"class_id"`
	Nis     string     `gorm:"type:varchar(30);not null;uniqueIndex:uq_students_nis" json:"nis"`
	Phone   *string    `gorm:"type:varchar(20)" json:"phone"`
	Address *string    `gorm:"type:varchar(255)" json:"address"`
	User    *User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Class   *Class     `gorm:"foreignKey:ClassID;references:ID" json:"class,omitempty"`
}

func (Student) TableName() string { return "students" }
