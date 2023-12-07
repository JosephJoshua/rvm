package auth

import (
	"context"
	"fmt"
)

var (
	ErrInvalidIDToken    = fmt.Errorf("invalid ID token")
	ErrGetUserFailed     = fmt.Errorf("failed to get user")
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
)

type Service struct {
	r  Repository
	ap AuthProvider
}

func NewService(r Repository, ap AuthProvider) *Service {
	return &Service{r: r, ap: ap}
}

func (s *Service) Register(idToken string) error {
	uid, err := s.ap.GetUserID(context.Background(), idToken)
	if err != nil {
		return fmt.Errorf("Register(): failed to get user ID: %w", ErrInvalidIDToken)
	}

	exists, err := s.r.DoesUserExist(uid)
	if err != nil {
		return fmt.Errorf("Register(): failed to check if user exists: %w", err)
	}

	if exists {
		return fmt.Errorf("Register(): user already exists: %w", ErrUserAlreadyExists)
	}

	info, err := s.ap.GetUserInfo(context.Background(), uid)
	if err != nil {
		return fmt.Errorf("Register(): failed to get user info: %w", ErrGetUserFailed)
	}

	err = s.r.CreateUser(uid, info.Name, info.Email)
	if err != nil {
		return fmt.Errorf("Register(): failed to create user: %w", err)
	}

	return nil
}
