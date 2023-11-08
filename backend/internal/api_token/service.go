package api_token

import (
	"errors"
	"fmt"
	"time"
)

type Service struct {
	r Repository
}

func NewService(r Repository) *Service {
	return &Service{
		r: r,
	}
}

func (s *Service) IsValidToken(tokenID string) (bool, error) {
	token, err := s.r.GetTokenByID(tokenID)

	if errors.Is(err, ErrTokenNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("IsValidToken(): failed to get token by id: %w", err)
	}

	if token.ExpiringAt != nil && token.ExpiringAt.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}
