package model

import (
	"github.com/satori/go.uuid"
)

//TokenInfo данные по токену
type TokenInfo struct {
	ClientID uuid.UUID
	UserID   uuid.UUID
}
