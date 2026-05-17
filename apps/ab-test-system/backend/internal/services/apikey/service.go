package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/apikey"
)

type Storage interface {
	Create(ctx context.Context, key domain.Key) (domain.Key, error)
	List(ctx context.Context, projectID uuid.UUID) ([]domain.Key, error)
	Revoke(ctx context.Context, id, projectID uuid.UUID) error
	GetByKeyHash(ctx context.Context, keyHash string) (domain.Key, error)
}

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(ctx context.Context, projectID uuid.UUID, name string) (domain.Key, string, error) {
	raw, err := generateRawKey()
	if err != nil {
		return domain.Key{}, "", fmt.Errorf("apikey.Service.Create: generate key: %w", err)
	}

	key := domain.Key{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      name,
		KeyHash:   sha256Hex(raw),
		Prefix:    raw[:7], // "sk_" + 4 hex chars
	}

	created, err := s.storage.Create(ctx, key)
	if err != nil {
		return domain.Key{}, "", fmt.Errorf("apikey.Service.Create: %w", err)
	}

	return created, raw, nil
}

func (s *Service) List(ctx context.Context, projectID uuid.UUID) ([]domain.Key, error) {
	keys, err := s.storage.List(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("apikey.Service.List: %w", err)
	}

	return keys, nil
}

func (s *Service) Revoke(ctx context.Context, projectID, keyID uuid.UUID) error {
	if err := s.storage.Revoke(ctx, keyID, projectID); err != nil {
		return fmt.Errorf("apikey.Service.Revoke: %w", err)
	}

	return nil
}

func (s *Service) Validate(ctx context.Context, rawKey string) (uuid.UUID, error) {
	key, err := s.storage.GetByKeyHash(ctx, sha256Hex(rawKey))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("apikey.Service.Validate: %w", err)
	}

	return key.ProjectID, nil
}

func generateRawKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return "sk_" + hex.EncodeToString(b), nil
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
