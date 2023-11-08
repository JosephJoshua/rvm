package apitoken

import (
	"errors"

	"github.com/JosephJoshua/rvm/backend/internal/apitoken/domain"
)

var (
	ErrTokenNotFound = errors.New("token not found")
)

type Repository interface {
	GetTokenByID(tokenID string) (*domain.APIToken, error)
}
