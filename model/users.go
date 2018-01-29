package model

import (
	"time"

	"github.com/satori/go.uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Name      string
	Login     string `gorm:"not null;unique"`
	Password  string
	Scopes    []Scope `gorm:"many2many:user_scopes;"`
}
