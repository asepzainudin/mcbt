package model

import (
	"github.com/google/uuid"
)

type QuestionBank struct {
	BaseModel
	SubjectID      uuid.UUID `gorm:"type:uuid;not null"`
	AcademicYearID *uuid.UUID
	CreatedBy      *uuid.UUID
	Title          string  `gorm:"type:varchar(150);not null;uniqueIndex:uq_question_banks_subject_title,priority:2"`
	Description    *string `gorm:"type:text"`

	Subject      Subject       `gorm:"foreignKey:SubjectID;references:ID"`
	AcademicYear *AcademicYear `gorm:"foreignKey:AcademicYearID;references:ID"`
	Creator      *User         `gorm:"foreignKey:CreatedBy;references:ID"`
	Questions    []Question    `gorm:"foreignKey:QuestionBankID;references:ID"`
}

func (QuestionBank) TableName() string { return "question_banks" }
