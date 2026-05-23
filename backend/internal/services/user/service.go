package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
)

type Storage interface {
	Upsert(ctx context.Context, u domain.User) (domain.User, error)
	GetByFirebaseUID(ctx context.Context, uid string) (domain.User, error)
}

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) CreateOrUpdate(ctx context.Context, p domain.UpsertParams) (domain.User, error) {
	u := domain.User{
		ID:          uuid.New(),
		FirebaseUID: p.FirebaseUID,
		Email:       p.Email,
		Name:        p.Name,
		PhotoURL:    p.PhotoURL,
		Role:        domain.RoleMember,
	}

	created, err := s.storage.Upsert(ctx, u)
	if err != nil {
		return domain.User{}, fmt.Errorf("user.Service.CreateOrUpdate: %w", err)
	}

	return created, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, firebaseUID string) (domain.User, error) {
	u, err := s.storage.GetByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return domain.User{}, fmt.Errorf("user.Service.GetCurrentUser: %w", err)
	}

	return u, nil
}
