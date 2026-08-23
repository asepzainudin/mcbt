package model

type Subject struct {
	BaseModel
	Code        string  `gorm:"type:varchar(20);not null;uniqueIndex:uq_subjects_code" json:"code"`
	Name        string  `gorm:"type:varchar(100);not null" json:"name"`
	Description *string `gorm:"type:text" json:"description"`

	QuestionBanks []QuestionBank `gorm:"foreignKey:SubjectID;references:ID" json:"-"`
}

func (Subject) TableName() string { return "subjects" }
