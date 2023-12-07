package auth

import "context"

type UserInfo struct {
	Email string
	Name  string
}

type AuthProvider interface {
	GetUserID(ctx context.Context, idToken string) (string, error)
	GetUserInfo(ctx context.Context, uid string) (*UserInfo, error)
}
