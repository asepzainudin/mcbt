package model

import (
	"time"

	"github.com/google/uuid"
)

type ExamAnswer struct {
	BaseModel
	AttemptID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_answers_attempt_question,priority:1;index:idx_answers_attempt" json:"attempt_id"`
	QuestionID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_answers_attempt_question,priority:2" json:"question_id"`
	AnswerValue     string    `gorm:"type:text;not null;default:''" json:"answer_value"`
	ClientTimestamp int64     `gorm:"not null;default:0" json:"client_timestamp"`
	IsFlagged       bool      `gorm:"not null;default:false" json:"is_flagged"`
	AnsweredAt      time.Time `gorm:"type:timestamptz;not null;default:now()" json:"answered_at"`

	Attempt  *ExamAttempt `gorm:"foreignKey:AttemptID;references:ID" json:"-"`
	Question *Question    `gorm:"foreignKey:QuestionID;references:ID" json:"-"`
}

func (ExamAnswer) TableName() string { return "exam_answers" }
