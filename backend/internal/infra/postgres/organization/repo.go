package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/organization"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, name string, userID uuid.UUID) (domain.Organization, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("Repo.Create: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var (
		orgID     uuid.UUID
		createdAt time.Time
	)
	err = tx.QueryRow(ctx,
		`INSERT INTO organizations (name) VALUES (@name) RETURNING id, created_at`,
		pgx.NamedArgs{"name": name},
	).Scan(&orgID, &createdAt)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("Repo.Create: insert org: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE users SET org_id = @org_id, role = 'admin' WHERE id = @user_id`,
		pgx.NamedArgs{"org_id": orgID, "user_id": userID},
	)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("Repo.Create: link user: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.Organization{}, fmt.Errorf("Repo.Create: commit: %w", err)
	}

	return domain.Organization{ID: orgID, Name: name, CreatedAt: createdAt}, nil
}

func (r *Repo) HasOrg(ctx context.Context, userID uuid.UUID) (bool, error) {
	var has bool
	err := r.db.QueryRow(ctx,
		`SELECT org_id IS NOT NULL FROM users WHERE id = @user_id`,
		pgx.NamedArgs{"user_id": userID},
	).Scan(&has)
	if err != nil {
		return false, fmt.Errorf("Repo.HasOrg: %w", err)
	}
	return has, nil
}
