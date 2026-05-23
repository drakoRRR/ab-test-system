package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/project"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type projectRow struct {
	ID          pgtype.UUID `db:"id"`
	OrgID       pgtype.UUID `db:"org_id"`
	Name        string      `db:"name"`
	Description string      `db:"description"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
}

func (r projectRow) toDomain() domain.Project {
	return domain.Project{
		ID:          uuid.UUID(r.ID.Bytes),
		OrgID:       uuid.UUID(r.OrgID.Bytes),
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func (r *Repo) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	const q = `
		INSERT INTO projects (id, org_id, name, description)
		VALUES (@id, @org_id, @name, @description)
		RETURNING id, org_id, name, description, created_at, updated_at`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{
		"id":          p.ID,
		"org_id":      p.OrgID,
		"name":        p.Name,
		"description": p.Description,
	})
	if err != nil {
		return domain.Project{}, fmt.Errorf("Repo.Create: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[projectRow])
	if err != nil {
		return domain.Project{}, fmt.Errorf("Repo.Create: %w", err)
	}

	return row.toDomain(), nil
}

func (r *Repo) List(ctx context.Context, orgID uuid.UUID) ([]domain.Project, error) {
	const q = `
		SELECT id, org_id, name, description, created_at, updated_at
		FROM projects WHERE org_id = @org_id ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"org_id": orgID})
	if err != nil {
		return nil, fmt.Errorf("Repo.List: %w", err)
	}

	projectRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[projectRow])
	if err != nil {
		return nil, fmt.Errorf("Repo.List: %w", err)
	}

	projects := make([]domain.Project, len(projectRows))
	for i, row := range projectRows {
		projects[i] = row.toDomain()
	}

	return projects, nil
}

func (r *Repo) GetByID(ctx context.Context, id, orgID uuid.UUID) (domain.Project, error) {
	const q = `
		SELECT id, org_id, name, description, created_at, updated_at
		FROM projects WHERE id = @id AND org_id = @org_id`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"id": id, "org_id": orgID})
	if err != nil {
		return domain.Project{}, fmt.Errorf("Repo.GetByID: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[projectRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, fmt.Errorf("Repo.GetByID: %w", domain.ErrNotFound)
		}

		return domain.Project{}, fmt.Errorf("Repo.GetByID: %w", err)
	}

	return row.toDomain(), nil
}

func (r *Repo) Update(ctx context.Context, p domain.Project) (domain.Project, error) {
	const q = `
		UPDATE projects SET name = @name, description = @description, updated_at = NOW()
		WHERE id = @id AND org_id = @org_id
		RETURNING id, org_id, name, description, created_at, updated_at`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{
		"id":          p.ID,
		"org_id":      p.OrgID,
		"name":        p.Name,
		"description": p.Description,
	})
	if err != nil {
		return domain.Project{}, fmt.Errorf("Repo.Update: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[projectRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, fmt.Errorf("Repo.Update: %w", domain.ErrNotFound)
		}

		return domain.Project{}, fmt.Errorf("Repo.Update: %w", err)
	}

	return row.toDomain(), nil
}

func (r *Repo) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	const q = `DELETE FROM projects WHERE id = @id AND org_id = @org_id`

	result, err := r.db.Exec(ctx, q, pgx.NamedArgs{"id": id, "org_id": orgID})
	if err != nil {
		return fmt.Errorf("Repo.Delete: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("Repo.Delete: %w", domain.ErrNotFound)
	}

	return nil
}
