package model

import (
	"github.com/google/uuid"
)

const (
	BankStatusDraft     = "draft"
	BankStatusPublished = "published"
	BankStatusArchived  = "archived"
)

type QuestionBank struct {
	BaseModel
	SubjectID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_question_banks_subject_title,priority:1" json:"subject_id"`
	AcademicYearID *uuid.UUID `gorm:"type:uuid;index:idx_question_banks_academic_year_id" json:"academic_year_id"`
	CreatedBy      *uuid.UUID `json:"-"`
	Code           string     `gorm:"type:varchar(50);not null;uniqueIndex:uq_question_banks_code" json:"code"`
	Status         string     `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	Title          string     `gorm:"type:varchar(150);not null;uniqueIndex:uq_question_banks_subject_title,priority:2" json:"title"`
	Description    *string    `gorm:"type:text" json:"description"`

	Subject      Subject       `gorm:"foreignKey:SubjectID;references:ID" json:"subject,omitempty"`
	AcademicYear *AcademicYear `gorm:"foreignKey:AcademicYearID;references:ID" json:"academic_year,omitempty"`
	Creator      *User         `gorm:"foreignKey:CreatedBy;references:ID" json:"-"`
	Questions    []Question    `gorm:"foreignKey:QuestionBankID;references:ID" json:"-"`
}

func (QuestionBank) TableName() string { return "question_banks" }
