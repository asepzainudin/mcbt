package model

import (
	"github.com/google/uuid"
)

const (
	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeTrueFalse      = "true_false"
	QuestionTypeMultipleAnswer = "multiple_answer"
	QuestionTypeEssay          = "essay"
	QuestionTypeShortAnswer    = "short_answer"
)

type Question struct {
	BaseModel
	QuestionBankID uuid.UUID  `gorm:"type:uuid;not null;index:idx_questions_bank_id" json:"question_bank_id"`
	MediaID        *uuid.UUID `gorm:"type:uuid;index:idx_questions_media_id" json:"media_id"`
	QuestionType   string     `gorm:"type:varchar(20);not null;default:'multiple_choice';index:idx_questions_type" json:"question_type"`
	Content        string     `gorm:"type:text;not null" json:"content"`
	ScoreWeight    float64    `gorm:"type:numeric(6,2);not null;default:1.0" json:"score_weight"`
	Explanation    *string    `gorm:"type:text" json:"explanation"`
	AnswerKeys     *string    `gorm:"type:text" json:"answer_keys,omitempty"`

	QuestionBank QuestionBank `gorm:"foreignKey:QuestionBankID;references:ID" json:"question_bank,omitempty"`
	Media        *Media       `gorm:"foreignKey:MediaID;references:ID" json:"media,omitempty"`
	Options      []Option     `gorm:"foreignKey:QuestionID;references:ID" json:"options,omitempty"`
}

func (Question) TableName() string { return "questions" }
