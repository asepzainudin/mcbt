package model

type Permission struct {
	BaseModel
	Name        string  `gorm:"type:varchar(100);not null;uniqueIndex:uq_permissions_name"`
	Description *string `gorm:"type:varchar(255)"`
	Roles       []Role  `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:permission_id;References:ID;joinReferences:role_id"`
}

func (Permission) TableName() string { return "permissions" }
