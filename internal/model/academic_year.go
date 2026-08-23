package model

import "time"

type AcademicYear struct {
	BaseModel
	Name      string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_academic_years_name"`
	StartDate time.Time `gorm:"type:date;not null"`
	EndDate   time.Time `gorm:"type:date;not null"`
	IsActive  bool      `gorm:"not null;default:false"`

	Classes       []Class        `gorm:"foreignKey:AcademicYearID;references:ID"`
	QuestionBanks []QuestionBank `gorm:"foreignKey:AcademicYearID;references:ID"`
}

func (AcademicYear) TableName() string { return "academic_years" }
