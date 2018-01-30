package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Client struct {
	ID        uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Secret    string
	Domain    string
	UserID    uuid.UUID
	Scope     []Scope `gorm:"many2many:client_scopes;"`
}
