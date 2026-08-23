package model

import (
	"github.com/google/uuid"
)

type Media struct {
	BaseModel
	UploadedBy *uuid.UUID `gorm:"type:uuid;index:idx_media_uploaded_by"`
	FileName   string     `gorm:"type:varchar(255);not null"`
	FilePath   string     `gorm:"type:varchar(500);not null;uniqueIndex:uq_media_file_path"`
	MimeType   string     `gorm:"type:varchar(100);not null"`
	FileSize   int64      `gorm:"not null;default:0"`

	Uploader  *User      `gorm:"foreignKey:UploadedBy;references:ID"`
	Questions []Question `gorm:"foreignKey:MediaID;references:ID"`
}

func (Media) TableName() string { return "media" }
