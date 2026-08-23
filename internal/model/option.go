package model

import (
	"github.com/google/uuid"
)

type Option struct {
	BaseModel
	QuestionID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_options_question_label,priority:1"`
	MediaID    *uuid.UUID `gorm:"type:uuid;index:idx_options_media_id"`
	Label      string     `gorm:"type:char(1);not null;uniqueIndex:uq_options_question_label,priority:2"`
	Content    string     `gorm:"type:text;not null"`
	IsCorrect  bool       `gorm:"not null;default:false"`

	Question Question `gorm:"foreignKey:QuestionID;references:ID"`
	Media    *Media   `gorm:"foreignKey:MediaID;references:ID"`
}

func (Option) TableName() string { return "question_options" }
