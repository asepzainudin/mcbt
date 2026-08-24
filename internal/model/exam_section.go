package model

import (
	"github.com/google/uuid"
)

type ExamSection struct {
	BaseModel
	ExamID   uuid.UUID `gorm:"type:uuid;not null;index:idx_exam_sections_exam_id" json:"exam_id"`
	Name     string    `gorm:"type:varchar(100);not null" json:"name"`
	Sequence int       `gorm:"not null" json:"sequence"`

	Exam      *Exam                 `gorm:"foreignKey:ExamID;references:ID" json:"exam,omitempty"`
	Questions []ExamSectionQuestion `gorm:"foreignKey:SectionID;references:ID" json:"questions,omitempty"`
}

func (ExamSection) TableName() string { return "exam_sections" }
