package model

import (
	"github.com/google/uuid"
)

type Class struct {
	BaseModel
	AcademicYearID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_classes_year_name,priority:1" json:"academic_year_id"`
	HomeroomTeacherID *uuid.UUID `gorm:"type:uuid;index:idx_classes_homeroom_teacher_id" json:"homeroom_teacher_id"`
	Name              string     `gorm:"type:varchar(100);not null;uniqueIndex:uq_classes_year_name,priority:2" json:"name"`
	GradeLevel        *int16     `gorm:"type:smallint" json:"grade_level"`

	AcademicYear    AcademicYear `gorm:"foreignKey:AcademicYearID;references:ID" json:"academic_year,omitempty"`
	HomeroomTeacher *Teacher     `gorm:"foreignKey:HomeroomTeacherID;references:ID" json:"-"`
	Students        []Student    `gorm:"foreignKey:ClassID;references:ID" json:"-"`
}

func (Class) TableName() string { return "classes" }
