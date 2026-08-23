package model

import (
	"github.com/google/uuid"
)

type Teacher struct {
	BaseModel
	UserID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_teachers_user_id" json:"user_id"`
	Nip     *string   `gorm:"type:varchar(30)" json:"nip"`
	Phone   *string   `gorm:"type:varchar(20)" json:"phone"`
	Address *string   `gorm:"type:varchar(255)" json:"address"`

	User            *User   `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	HomeroomClasses []Class `gorm:"foreignKey:HomeroomTeacherID;references:ID" json:"-"`
}

func (Teacher) TableName() string { return "teachers" }
