package flag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type flagRow struct {
	ID        pgtype.UUID `db:"id"`
	ProjectID pgtype.UUID `db:"project_id"`
	Key       string      `db:"key"`
	Name      string      `db:"name"`
	Enabled   bool        `db:"enabled"`
	Rules     []byte      `db:"rules"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

func (r flagRow) toDomain() (domain.Flag, error) {
	var rules []domain.Rule
	if err := json.Unmarshal(r.Rules, &rules); err != nil {
		return domain.Flag{}, fmt.Errorf("unmarshal rules: %w", err)
	}

	return domain.Flag{
		ID:        uuid.UUID(r.ID.Bytes),
		ProjectID: uuid.UUID(r.ProjectID.Bytes),
		Key:       r.Key,
		Name:      r.Name,
		Enabled:   r.Enabled,
		Rules:     rules,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}, nil
}

func marshalRules(rules []domain.Rule) ([]byte, error) {
	if rules == nil {
		rules = []domain.Rule{}
	}

	return json.Marshal(rules)
}

func (r *Repo) Create(ctx context.Context, f domain.Flag) (domain.Flag, error) {
	rulesJSON, err := marshalRules(f.Rules)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("Repo.Create: %w", err)
	}

	const q = `
		INSERT INTO flags (id, project_id, key, name, enabled, rules)
		VALUES (@id, @project_id, @key, @name, @enabled, @rules)
		RETURNING id, project_id, key, name, enabled, rules, created_at, updated_at`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{
		"id":         f.ID,
		"project_id": f.ProjectID,
		"key":        f.Key,
		"name":       f.Name,
		"enabled":    f.Enabled,
		"rules":      rulesJSON,
	})
	if err != nil {
		return domain.Flag{}, fmt.Errorf("Repo.Create: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[flagRow])
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Flag{}, fmt.Errorf("Repo.Create: %w", domain.ErrConflict)
		}

		return domain.Flag{}, fmt.Errorf("Repo.Create: %w", err)
	}

	return row.toDomain()
}

func (r *Repo) List(ctx context.Context, projectID uuid.UUID) ([]domain.Flag, error) {
	const q = `
		SELECT id, project_id, key, name, enabled, rules, created_at, updated_at
		FROM flags WHERE project_id = @project_id ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"project_id": projectID})
	if err != nil {
		return nil, fmt.Errorf("Repo.List: %w", err)
	}

	flagRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[flagRow])
	if err != nil {
		return nil, fmt.Errorf("Repo.List: %w", err)
	}

	flags := make([]domain.Flag, len(flagRows))
	for i, row := range flagRows {
		f, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("Repo.List: %w", err)
		}

		flags[i] = f
	}

	return flags, nil
}

func (r *Repo) GetByKey(ctx context.Context, projectID uuid.UUID, key string) (domain.Flag, error) {
	const q = `
		SELECT id, project_id, key, name, enabled, rules, created_at, updated_at
		FROM flags WHERE project_id = @project_id AND key = @key`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"project_id": projectID, "key": key})
	if err != nil {
		return domain.Flag{}, fmt.Errorf("Repo.GetByKey: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[flagRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Flag{}, fmt.Errorf("Repo.GetByKey: %w", domain.ErrNotFound)
		}

		return domain.Flag{}, fmt.Errorf("Repo.GetByKey: %w", err)
	}

	return row.toDomain()
}

func (r *Repo) Update(ctx context.Context, f domain.Flag) (domain.Flag, error) {
	rulesJSON, err := marshalRules(f.Rules)
	if err != nil {
		return domain.Flag{}, fmt.Errorf("Repo.Update: %w", err)
	}

	const q = `
		UPDATE flags SET name = @name, enabled = @enabled, rules = @rules, updated_at = NOW()
		WHERE id = @id AND project_id = @project_id
		RETURNING id, project_id, key, name, enabled, rules, created_at, updated_at`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{
		"id":         f.ID,
		"project_id": f.ProjectID,
		"name":       f.Name,
		"enabled":    f.Enabled,
		"rules":      rulesJSON,
	})
	if err != nil {
		return domain.Flag{}, fmt.Errorf("Repo.Update: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[flagRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Flag{}, fmt.Errorf("Repo.Update: %w", domain.ErrNotFound)
		}

		return domain.Flag{}, fmt.Errorf("Repo.Update: %w", err)
	}

	return row.toDomain()
}

func (r *Repo) Delete(ctx context.Context, projectID uuid.UUID, key string) error {
	const q = `DELETE FROM flags WHERE project_id = @project_id AND key = @key`

	result, err := r.db.Exec(ctx, q, pgx.NamedArgs{"project_id": projectID, "key": key})
	if err != nil {
		return fmt.Errorf("Repo.Delete: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("Repo.Delete: %w", domain.ErrNotFound)
	}

	return nil
}

// isUniqueViolation checks for PostgreSQL unique constraint violation (code 23505).
func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}

	return false
}
