package model

import (
	"github.com/satori/go.uuid"
)

type UserScopes struct {
	UserID  uuid.UUID
	ScopeID uuid.UUID
}
