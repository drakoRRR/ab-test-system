package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	projectIDKey contextKey = "projectID"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (string, error)
}

type ApiKeyVerifier interface {
	Validate(ctx context.Context, rawKey string) (uuid.UUID, error)
}

func Auth(v TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"message":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			token, found := strings.CutPrefix(authHeader, "Bearer ")
			if !found {
				http.Error(w, `{"message":"authorization header must be Bearer <token>"}`, http.StatusUnauthorized)
				return
			}

			uid, err := v.VerifyToken(r.Context(), token)
			if err != nil {
				log.Error().Err(err).Str("path", r.URL.Path).Msg("token verification failed")
				http.Error(w, `{"message":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ApiKeyAuth(v ApiKeyVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := r.Header.Get("X-API-Key")
			if rawKey == "" {
				http.Error(w, `{"message":"missing X-API-Key header"}`, http.StatusUnauthorized)
				return
			}

			projectID, err := v.Validate(r.Context(), rawKey)
			if err != nil {
				http.Error(w, `{"message":"invalid or revoked API key"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), projectIDKey, projectID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(userIDKey).(string)
	return uid, ok
}

func ContextWithUserID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func ProjectIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(projectIDKey).(uuid.UUID)
	return id, ok
}

func ContextWithProjectID(ctx context.Context, projectID uuid.UUID) context.Context {
	return context.WithValue(ctx, projectIDKey, projectID)
}
