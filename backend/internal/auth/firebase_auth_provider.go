package auth

import (
	"context"
	"fmt"

	"github.com/JosephJoshua/rvm/backend/internal/firebase"
)

type FirebaseAuthProvider struct {
	app *firebase.App
}

func NewFirebaseAuthProvider(app *firebase.App) *FirebaseAuthProvider {
	return &FirebaseAuthProvider{app: app}
}

func (p *FirebaseAuthProvider) GetUserID(ctx context.Context, idToken string) (string, error) {
	token, err := p.app.Auth().VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", fmt.Errorf("VerifyIDToken(): failed to verify ID token: %w", err)
	}

	return token.UID, nil
}

func (p *FirebaseAuthProvider) GetUserInfo(ctx context.Context, uid string) (*UserInfo, error) {
	user, err := p.app.Auth().GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("GetUserInfo(): failed to get user: %w", err)
	}

	return &UserInfo{
		Email: user.Email,
		Name:  user.DisplayName,
	}, nil
}
