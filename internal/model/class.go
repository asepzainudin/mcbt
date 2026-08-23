package model

import (
	"github.com/google/uuid"
)

type Class struct {
	BaseModel
	AcademicYearID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_classes_year_name,priority:1"`
	HomeroomTeacherID *uuid.UUID `gorm:"type:uuid;index:idx_classes_homeroom_teacher_id"`
	Name              string     `gorm:"type:varchar(100);not null;uniqueIndex:uq_classes_year_name,priority:2"`
	GradeLevel        *int16     `gorm:"type:smallint"`

	AcademicYear    AcademicYear `gorm:"foreignKey:AcademicYearID;references:ID"`
	HomeroomTeacher *Teacher     `gorm:"foreignKey:HomeroomTeacherID;references:ID"`
	Students        []Student    `gorm:"foreignKey:ClassID;references:ID"`
}

func (Class) TableName() string { return "classes" }
