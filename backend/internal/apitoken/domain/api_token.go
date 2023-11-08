package domain

import (
	"time"
)

type APIToken struct {
	ID         string
	ExpiringAt *time.Time
	CreatedAt  time.Time
}

func NewAPIToken(id string, expiringAt *time.Time, createdAt time.Time) *APIToken {
	return &APIToken{
		ID:         id,
		ExpiringAt: expiringAt,
		CreatedAt:  createdAt,
	}
}
