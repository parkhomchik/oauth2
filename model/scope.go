package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Scope struct {
	ID        uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Name      string `gorm:"not null;unique"`
}
