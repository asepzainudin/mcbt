package model

import (
	"github.com/google/uuid"
)

type ExamSectionQuestion struct {
	BaseModel
	SectionID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_section_questions,priority:1;index:idx_section_questions_section" json:"section_id"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_section_questions,priority:2;index:idx_section_questions_question" json:"question_id"`

	Section  *ExamSection `gorm:"foreignKey:SectionID;references:ID" json:"-"`
	Question *Question    `gorm:"foreignKey:QuestionID;references:ID" json:"question,omitempty"`
}

func (ExamSectionQuestion) TableName() string { return "exam_section_questions" }
