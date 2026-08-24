package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	AttemptStatusInProgress = "in_progress"
	AttemptStatusSubmitted  = "submitted"
	AttemptStatusExpired    = "expired"
)

type ExamAttempt struct {
	BaseModel
	ExamID      uuid.UUID  `gorm:"type:uuid;not null;index:idx_attempts_exam_student" json:"exam_id"`
	StudentID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_attempts_student" json:"student_id"`
	AttemptNo   int        `gorm:"not null" json:"attempt_no"`
	Status      string     `gorm:"type:varchar(20);not null;default:'in_progress'" json:"status"`
	StartedAt   time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"started_at"`
	ExpiresAt   time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	SubmittedAt *time.Time `gorm:"type:timestamptz" json:"submitted_at"`
	Score       *float64   `gorm:"type:numeric(5,2)" json:"score"`

	Exam    *Exam    `gorm:"foreignKey:ExamID;references:ID" json:"exam,omitempty"`
	Student *Student `gorm:"foreignKey:StudentID;references:ID" json:"-"`
}

func (ExamAttempt) TableName() string { return "exam_attempts" }
