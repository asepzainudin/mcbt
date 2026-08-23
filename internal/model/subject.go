package model

type Subject struct {
	BaseModel
	Code        string  `gorm:"type:varchar(20);not null;uniqueIndex:uq_subjects_code"`
	Name        string  `gorm:"type:varchar(100);not null"`
	Description *string `gorm:"type:text"`

	QuestionBanks []QuestionBank `gorm:"foreignKey:SubjectID;references:ID"`
}

func (Subject) TableName() string { return "subjects" }
