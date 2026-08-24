package model

import (
	"github.com/google/uuid"
)

const (
	AssignedViaClass      = "class"
	AssignedViaIndividual = "individual"
)

type ExamParticipant struct {
	BaseModel
	ExamID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_exam_participants,priority:1;index:idx_exam_participants_exam" json:"exam_id"`
	StudentID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_exam_participants,priority:2;index:idx_exam_participants_student" json:"student_id"`
	AssignedVia string    `gorm:"type:varchar(10);not null;default:'individual'" json:"assigned_via"`

	Exam    *Exam    `gorm:"foreignKey:ExamID;references:ID" json:"-"`
	Student *Student `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"`
}

func (ExamParticipant) TableName() string { return "exam_participants" }
