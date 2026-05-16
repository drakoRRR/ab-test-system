package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/user"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type userRow struct {
	ID          pgtype.UUID `db:"id"`
	FirebaseUID string      `db:"firebase_uid"`
	OrgID       pgtype.UUID `db:"org_id"`
	Email       string      `db:"email"`
	Name        string      `db:"name"`
	PhotoURL    *string     `db:"photo_url"`
	Role        string      `db:"role"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
}

func (r userRow) toDomain() domain.User {
	u := domain.User{
		ID:          uuid.UUID(r.ID.Bytes),
		FirebaseUID: r.FirebaseUID,
		Email:       r.Email,
		Name:        r.Name,
		PhotoURL:    r.PhotoURL,
		Role:        domain.Role(r.Role),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if r.OrgID.Valid {
		orgID := uuid.UUID(r.OrgID.Bytes)
		u.OrgID = &orgID
	}

	return u
}

func (r *Repo) Upsert(ctx context.Context, u domain.User) (domain.User, error) {
	const q = `
		INSERT INTO users (id, firebase_uid, email, name, photo_url, role)
		VALUES (@id, @firebase_uid, @email, @name, @photo_url, @role)
		ON CONFLICT (firebase_uid) DO UPDATE SET
			email      = EXCLUDED.email,
			name       = EXCLUDED.name,
			photo_url  = EXCLUDED.photo_url,
			updated_at = NOW()
		RETURNING id, firebase_uid, org_id, email, name, photo_url, role, created_at, updated_at`

	args := pgx.NamedArgs{
		"id":           u.ID,
		"firebase_uid": u.FirebaseUID,
		"email":        u.Email,
		"name":         u.Name,
		"photo_url":    u.PhotoURL,
		"role":         string(u.Role),
	}

	rows, err := r.db.Query(ctx, q, args)
	if err != nil {
		return domain.User{}, fmt.Errorf("Repo.Upsert: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		return domain.User{}, fmt.Errorf("Repo.Upsert: %w", err)
	}

	return row.toDomain(), nil
}

func (r *Repo) GetByFirebaseUID(ctx context.Context, uid string) (domain.User, error) {
	const q = `
		SELECT id, firebase_uid, org_id, email, name, photo_url, role, created_at, updated_at
		FROM users WHERE firebase_uid = @firebase_uid`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"firebase_uid": uid})
	if err != nil {
		return domain.User{}, fmt.Errorf("Repo.GetByFirebaseUID: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("Repo.GetByFirebaseUID: %w", domain.ErrNotFound)
		}

		return domain.User{}, fmt.Errorf("Repo.GetByFirebaseUID: %w", err)
	}

	return row.toDomain(), nil
}
