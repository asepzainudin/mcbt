package model

import (
	"github.com/google/uuid"
)

const (
	ExamStatusDraft     = "draft"
	ExamStatusPublished = "published"
	ExamStatusClosed    = "closed"
)

type Exam struct {
	BaseModel
	Title                 string     `gorm:"type:varchar(150);not null" json:"title"`
	Description           *string    `gorm:"type:text" json:"description"`
	SubjectID             uuid.UUID  `gorm:"type:uuid;not null;index:idx_exams_subject_id" json:"subject_id"`
	AcademicYearID        *uuid.UUID `gorm:"type:uuid;index:idx_exams_academic_year_id" json:"academic_year_id"`
	CreatedBy             *uuid.UUID `gorm:"type:uuid;index:idx_exams_created_by" json:"-"`
	QuestionBankID        *uuid.UUID `gorm:"type:uuid;index:idx_exams_question_bank_id" json:"question_bank_id"`
	Status                string     `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	DurationMinutes       int        `gorm:"not null;default:60" json:"duration_minutes"`
	MaxAttempts           int        `gorm:"not null;default:1" json:"max_attempts"`
	PassingGrade          float64    `gorm:"type:numeric(5,2);not null;default:75.00" json:"passing_grade"`
	RandomizeQuestions    bool       `gorm:"not null;default:false" json:"randomize_questions"`
	RandomizeOptions      bool       `gorm:"not null;default:false" json:"randomize_options"`
	AllowBacktrack        bool       `gorm:"not null;default:true" json:"allow_backtrack"`
	AutoSubmit            bool       `gorm:"not null;default:true" json:"auto_submit"`
	ShowResultImmediately bool       `gorm:"not null;default:false" json:"show_result_immediately"`
	NegativeMarking       bool       `gorm:"not null;default:false" json:"negative_marking"`
	NegativeValue         float64    `gorm:"type:numeric(4,2);not null;default:0.00" json:"negative_value"`
	TokenEnabled          bool       `gorm:"not null;default:false" json:"token_enabled"`
	ResultsPublished      bool       `gorm:"not null;default:false" json:"results_published"`
	AllowDiscussion       bool       `gorm:"not null;default:false" json:"allow_discussion"`
	ExamToken             *string    `gorm:"type:varchar(10);uniqueIndex:uq_exams_token" json:"exam_token,omitempty"`

	AttemptsCount int64 `gorm:"-" json:"attempts_count"`

	Subject      Subject       `gorm:"foreignKey:SubjectID;references:ID" json:"subject,omitempty"`
	AcademicYear *AcademicYear `gorm:"foreignKey:AcademicYearID;references:ID" json:"academic_year,omitempty"`
	QuestionBank *QuestionBank `gorm:"foreignKey:QuestionBankID;references:ID" json:"question_bank,omitempty"`
}

func (Exam) TableName() string { return "exams" }
