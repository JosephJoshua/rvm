package api_token

import (
	"errors"

	"github.com/JosephJoshua/rvm/backend/internal/api_token/domain"
)

var (
	ErrTokenNotFound = errors.New("token not found")
)

type Repository interface {
	GetTokenByID(tokenID string) (*domain.APIToken, error)
}
