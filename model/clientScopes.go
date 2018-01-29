package model

import (
	"github.com/satori/go.uuid"
)

type ClientScopes struct {
	ClientID uuid.UUID
	ScopeID  uuid.UUID
}
