package model

import (
	"github.com/google/uuid"
)

type Student struct {
	BaseModel
	UserID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_students_user_id"`
	ClassID *uuid.UUID `gorm:"type:uuid;index:idx_students_class_id"`
	Nis     string     `gorm:"type:varchar(30);not null;uniqueIndex:uq_students_nis"`
	Phone   *string    `gorm:"type:varchar(20)"`
	Address *string    `gorm:"type:varchar(255)"`

	User  *User  `gorm:"foreignKey:UserID;references:ID"`
	Class *Class `gorm:"foreignKey:ClassID;references:ID"`
}

func (Student) TableName() string { return "students" }
