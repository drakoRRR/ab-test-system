package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"

	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/infra/gcp"
)

type Auth struct {
	client *auth.Client
}

func New(ctx context.Context, gcpConf gcp.Config) (*Auth, error) {
	c := &firebase.Config{ProjectID: gcpConf.ProjectID}

	app, err := firebase.NewApp(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("initialising firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating  firebase auth client: %w", err)
	}

	return &Auth{client: client}, nil
}

func (a *Auth) VerifyToken(ctx context.Context, token string) (string, error) {
	t, err := a.client.VerifyIDToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("verifying firebase auth token: %w", err)
	}

	return t.UID, nil
}
