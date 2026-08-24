package model

import (
	"github.com/google/uuid"
)

type Option struct {
	BaseModel
	QuestionID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_options_question_label,priority:1" json:"question_id"`
	MediaID    *uuid.UUID `gorm:"type:uuid;index:idx_options_media_id" json:"media_id"`
	Label      string     `gorm:"type:char(1);not null;uniqueIndex:uq_options_question_label,priority:2" json:"label"`
	Content    string     `gorm:"type:text;not null" json:"content"`
	IsCorrect  bool       `gorm:"not null;default:false" json:"is_correct"`
	Position   int16      `gorm:"type:smallint;not null;default:0" json:"position"`

	Question Question `gorm:"foreignKey:QuestionID;references:ID" json:"-"`
	Media    *Media   `gorm:"foreignKey:MediaID;references:ID" json:"media,omitempty"`
}

func (Option) TableName() string { return "question_options" }
