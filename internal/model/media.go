package model

import (
	"github.com/google/uuid"
)

type Media struct {
	BaseModel
	UploadedBy *uuid.UUID `gorm:"type:uuid;index:idx_media_uploaded_by" json:"uploaded_by"`
	FileName   string     `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath   string     `gorm:"type:varchar(500);not null;uniqueIndex:uq_media_file_path" json:"file_path"`
	MimeType   string     `gorm:"type:varchar(100);not null" json:"mime_type"`
	FileSize   int64      `gorm:"not null;default:0" json:"file_size"`

	Uploader  *User      `gorm:"foreignKey:UploadedBy;references:ID" json:"-"`
	Questions []Question `gorm:"foreignKey:MediaID;references:ID" json:"-"`
}

func (Media) TableName() string { return "media" }
