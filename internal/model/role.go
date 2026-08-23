package model

type Role struct {
	BaseModel
	Name        string       `gorm:"type:varchar(50);not null;uniqueIndex:uq_roles_name"`
	Description *string      `gorm:"type:varchar(255)"`
	Permissions []Permission `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:role_id;References:ID;joinReferences:permission_id"`
	Users       []User       `gorm:"foreignKey:RoleID;references:ID"`
}

func (Role) TableName() string { return "roles" }
