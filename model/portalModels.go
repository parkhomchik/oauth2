package model

import uuid "github.com/satori/go.uuid"

type Staff struct {
	ID       uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Name     string    `gorm:"type:varchar;"`
	RoleID   uuid.UUID `gorm:"type:uuid REFERENCES Role(Id)"`
	ParentID uuid.UUID //ID владельца компании
	UserID   uuid.UUID `json:"-"` //Связь с oauth
}

func (Staff) TableName() string {
	return "staff"
}

type Role struct {
	ID   uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Name string    `gorm:"type:varchar;"`
}

func (Role) TableName() string {
	return "role"
}
