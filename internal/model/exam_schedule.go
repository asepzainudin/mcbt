package model

import (
	"time"

	"github.com/google/uuid"
)

type ExamSchedule struct {
	BaseModel
	ExamID    uuid.UUID `gorm:"type:uuid;not null" json:"exam_id"`
	StartTime time.Time `gorm:"type:timestamptz;not null" json:"start_time"`
	EndTime   time.Time `gorm:"type:timestamptz;not null" json:"end_time"`
	Token     string    `gorm:"type:varchar(10);not null;uniqueIndex:uq_exam_schedules_token" json:"token"`

	Exam *Exam `gorm:"foreignKey:ExamID;references:ID" json:"exam,omitempty"`
}

func (ExamSchedule) TableName() string { return "exam_schedules" }
