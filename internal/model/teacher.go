package model

import (
	"github.com/google/uuid"
)

type Teacher struct {
	BaseModel
	UserID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_teachers_user_id"`
	Nip     *string   `gorm:"type:varchar(30)"`
	Phone   *string   `gorm:"type:varchar(20)"`
	Address *string   `gorm:"type:varchar(255)"`

	User            *User   `gorm:"foreignKey:UserID;references:ID"`
	HomeroomClasses []Class `gorm:"foreignKey:HomeroomTeacherID;references:ID"`
}

func (Teacher) TableName() string { return "teachers" }
