package model

import "time"

type AcademicYear struct {
	BaseModel
	Year      string     `gorm:"type:varchar(20);not null;uniqueIndex:uq_academic_years_year_semester,priority:1" json:"year"`
	Semester  string     `gorm:"type:varchar(9);not null;uniqueIndex:uq_academic_years_year_semester,priority:2;default:'ODD'" json:"semester"`
	StartDate *time.Time `gorm:"type:date" json:"start_date"`
	EndDate   *time.Time `gorm:"type:date" json:"end_date"`
	IsActive  bool       `gorm:"not null;default:false" json:"is_active"`

	Classes       []Class        `gorm:"foreignKey:AcademicYearID;references:ID" json:"-"`
	QuestionBanks []QuestionBank `gorm:"foreignKey:AcademicYearID;references:ID" json:"-"`
}

func (AcademicYear) TableName() string { return "academic_years" }
