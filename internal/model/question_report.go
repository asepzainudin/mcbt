package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	ReportStatusPending   = "pending"
	ReportStatusReviewing = "reviewing"
	ReportStatusResolved  = "resolved"
	ReportStatusRejected  = "rejected"
)

type QuestionReport struct {
	BaseModel
	AttemptID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_reports_attempt_question,priority:1;index:idx_reports_question" json:"attempt_id"`
	QuestionID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_reports_attempt_question,priority:2;index:idx_reports_question" json:"question_id"`
	StudentID  uuid.UUID  `gorm:"type:uuid;not null" json:"student_id"`
	Reason     string     `gorm:"type:text;not null" json:"reason"`
	Status     string     `gorm:"type:varchar(20);not null;default:'pending';index:idx_reports_status" json:"status"`
	Resolution *string    `gorm:"type:text" json:"resolution,omitempty"`
	ResolvedBy *uuid.UUID `gorm:"type:uuid" json:"resolved_by,omitempty"`
	ResolvedAt *time.Time `gorm:"type:timestamptz" json:"resolved_at,omitempty"`

	Attempt  *ExamAttempt `gorm:"foreignKey:AttemptID;references:ID" json:"attempt,omitempty"`
	Question *Question    `gorm:"foreignKey:QuestionID;references:ID" json:"question,omitempty"`
	Student  *Student     `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"`
	Resolver *User        `gorm:"foreignKey:ResolvedBy;references:ID" json:"resolver,omitempty"`
}

func (QuestionReport) TableName() string { return "question_reports" }
