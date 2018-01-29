package model

import uuid "github.com/satori/go.uuid"

type Staff struct {
	Id   uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	Name string    `gorm:"type:varchar;"`
	//Password string    `gorm:"type:varchar;"`
	Role     uuid.UUID `gorm:"type:uuid REFERENCES Role(Id)"` // тут нужно подумать, возможно будет несколько ролей
	ParentId uuid.UUID
	Visible  bool `gorm:"type:bool;"`
}
