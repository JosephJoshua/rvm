package firebase

import (
	"context"
	"fmt"

	f "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

type App struct {
	app  *f.App
	auth *auth.Client
}

func NewApp(ctx context.Context, credentialsJSON []byte) (*App, error) {
	opt := option.WithCredentialsJSON(credentialsJSON)
	app, err := f.NewApp(ctx, nil, opt)

	if err != nil {
		return nil, fmt.Errorf("NewFirebaseApp(): failed to create new app: %w", err)
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewFirebaseApp(): failed to create new client: %w", err)
	}

	return &App{app: app, auth: auth}, nil
}

func (a *App) Auth() *auth.Client {
	return a.auth
}
