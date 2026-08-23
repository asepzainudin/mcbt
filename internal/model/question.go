package model

import (
	"github.com/google/uuid"
)

const (
	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeTrueFalse      = "true_false"
	QuestionTypeEssay          = "essay"
)

type Question struct {
	BaseModel
	QuestionBankID uuid.UUID  `gorm:"type:uuid;not null;index:idx_questions_bank_id"`
	MediaID        *uuid.UUID `gorm:"type:uuid;index:idx_questions_media_id"`
	QuestionType   string     `gorm:"type:varchar(20);not null;default:'multiple_choice';index:idx_questions_type"`
	Content        string     `gorm:"type:text;not null"`
	Points         int        `gorm:"not null;default:1"`
	Explanation    *string    `gorm:"type:text"`

	QuestionBank QuestionBank `gorm:"foreignKey:QuestionBankID;references:ID"`
	Media        *Media       `gorm:"foreignKey:MediaID;references:ID"`
	Options      []Option     `gorm:"foreignKey:QuestionID;references:ID"`
}

func (Question) TableName() string { return "questions" }
