package apikey

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/apikey"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type keyRow struct {
	ID        pgtype.UUID `db:"id"`
	ProjectID pgtype.UUID `db:"project_id"`
	Name      string      `db:"name"`
	KeyHash   string      `db:"key_hash"`
	Prefix    string      `db:"prefix"`
	CreatedAt time.Time   `db:"created_at"`
	RevokedAt *time.Time  `db:"revoked_at"`
}

func (r keyRow) toDomain() domain.Key {
	return domain.Key{
		ID:        uuid.UUID(r.ID.Bytes),
		ProjectID: uuid.UUID(r.ProjectID.Bytes),
		Name:      r.Name,
		KeyHash:   r.KeyHash,
		Prefix:    r.Prefix,
		CreatedAt: r.CreatedAt,
		RevokedAt: r.RevokedAt,
	}
}

func (r *Repo) Create(ctx context.Context, key domain.Key) (domain.Key, error) {
	const q = `
		INSERT INTO sdk_keys (id, project_id, name, key_hash, prefix)
		VALUES (@id, @project_id, @name, @key_hash, @prefix)
		RETURNING id, project_id, name, key_hash, prefix, created_at, revoked_at`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{
		"id":         key.ID,
		"project_id": key.ProjectID,
		"name":       key.Name,
		"key_hash":   key.KeyHash,
		"prefix":     key.Prefix,
	})
	if err != nil {
		return domain.Key{}, fmt.Errorf("Repo.Create: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[keyRow])
	if err != nil {
		return domain.Key{}, fmt.Errorf("Repo.Create: %w", err)
	}

	return row.toDomain(), nil
}

func (r *Repo) List(ctx context.Context, projectID uuid.UUID) ([]domain.Key, error) {
	const q = `
		SELECT id, project_id, name, key_hash, prefix, created_at, revoked_at
		FROM sdk_keys WHERE project_id = @project_id ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"project_id": projectID})
	if err != nil {
		return nil, fmt.Errorf("Repo.List: %w", err)
	}

	keyRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[keyRow])
	if err != nil {
		return nil, fmt.Errorf("Repo.List: %w", err)
	}

	keys := make([]domain.Key, len(keyRows))
	for i, row := range keyRows {
		keys[i] = row.toDomain()
	}

	return keys, nil
}

func (r *Repo) Revoke(ctx context.Context, id, projectID uuid.UUID) error {
	const q = `
		UPDATE sdk_keys SET revoked_at = NOW()
		WHERE id = @id AND project_id = @project_id AND revoked_at IS NULL`

	result, err := r.db.Exec(ctx, q, pgx.NamedArgs{"id": id, "project_id": projectID})
	if err != nil {
		return fmt.Errorf("Repo.Revoke: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("Repo.Revoke: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *Repo) GetByKeyHash(ctx context.Context, keyHash string) (domain.Key, error) {
	const q = `
		SELECT id, project_id, name, key_hash, prefix, created_at, revoked_at
		FROM sdk_keys WHERE key_hash = @key_hash AND revoked_at IS NULL`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"key_hash": keyHash})
	if err != nil {
		return domain.Key{}, fmt.Errorf("Repo.GetByKeyHash: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[keyRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Key{}, fmt.Errorf("Repo.GetByKeyHash: %w", domain.ErrNotFound)
		}

		return domain.Key{}, fmt.Errorf("Repo.GetByKeyHash: %w", err)
	}

	return row.toDomain(), nil
}
