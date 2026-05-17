package experiment

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

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type experimentRow struct {
	ID             pgtype.UUID `db:"id"`
	ProjectID      pgtype.UUID `db:"project_id"`
	FlagID         pgtype.UUID `db:"flag_id"`
	Name           string      `db:"name"`
	Description    string      `db:"description"`
	Status         string      `db:"status"`
	TrafficPercent float64     `db:"traffic_percent"`
	Variants       []byte      `db:"variants"`
	CreatedAt      time.Time   `db:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at"`
	StartedAt      *time.Time  `db:"started_at"`
	EndedAt        *time.Time  `db:"ended_at"`
}

func (r experimentRow) toDomain() (domain.Experiment, error) {
	var variants []domain.Variant
	if err := json.Unmarshal(r.Variants, &variants); err != nil {
		return domain.Experiment{}, fmt.Errorf("unmarshal variants: %w", err)
	}

	exp := domain.Experiment{
		ID:             uuid.UUID(r.ID.Bytes),
		ProjectID:      uuid.UUID(r.ProjectID.Bytes),
		Name:           r.Name,
		Description:    r.Description,
		Status:         domain.Status(r.Status),
		TrafficPercent: r.TrafficPercent,
		Variants:       variants,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		StartedAt:      r.StartedAt,
		EndedAt:        r.EndedAt,
	}

	if r.FlagID.Valid {
		id := uuid.UUID(r.FlagID.Bytes)
		exp.FlagID = &id
	}

	return exp, nil
}

func marshalVariants(variants []domain.Variant) ([]byte, error) {
	if variants == nil {
		variants = []domain.Variant{}
	}

	return json.Marshal(variants)
}

func (r *Repo) Create(ctx context.Context, exp domain.Experiment) (domain.Experiment, error) {
	variantsJSON, err := marshalVariants(exp.Variants)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("Repo.Create: %w", err)
	}

	var flagID *uuid.UUID
	if exp.FlagID != nil {
		flagID = exp.FlagID
	}

	const q = `
		INSERT INTO experiments (id, project_id, flag_id, name, description, status, traffic_percent, variants)
		VALUES (@id, @project_id, @flag_id, @name, @description, @status, @traffic_percent, @variants)
		RETURNING id, project_id, flag_id, name, description, status, traffic_percent, variants,
		          created_at, updated_at, started_at, ended_at`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{
		"id":              exp.ID,
		"project_id":      exp.ProjectID,
		"flag_id":         flagID,
		"name":            exp.Name,
		"description":     exp.Description,
		"status":          string(exp.Status),
		"traffic_percent": exp.TrafficPercent,
		"variants":        variantsJSON,
	})
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("Repo.Create: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[experimentRow])
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Experiment{}, fmt.Errorf("Repo.Create: %w", domain.ErrConflict)
		}

		return domain.Experiment{}, fmt.Errorf("Repo.Create: %w", err)
	}

	return row.toDomain()
}

func (r *Repo) List(ctx context.Context, projectID uuid.UUID) ([]domain.Experiment, error) {
	const q = `
		SELECT id, project_id, flag_id, name, description, status, traffic_percent, variants,
		       created_at, updated_at, started_at, ended_at
		FROM experiments WHERE project_id = @project_id ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"project_id": projectID})
	if err != nil {
		return nil, fmt.Errorf("Repo.List: %w", err)
	}

	expRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[experimentRow])
	if err != nil {
		return nil, fmt.Errorf("Repo.List: %w", err)
	}

	experiments := make([]domain.Experiment, len(expRows))
	for i, row := range expRows {
		exp, err := row.toDomain()
		if err != nil {
			return nil, fmt.Errorf("Repo.List: %w", err)
		}

		experiments[i] = exp
	}

	return experiments, nil
}

func (r *Repo) GetByID(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error) {
	const q = `
		SELECT id, project_id, flag_id, name, description, status, traffic_percent, variants,
		       created_at, updated_at, started_at, ended_at
		FROM experiments WHERE project_id = @project_id AND id = @id`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{"project_id": projectID, "id": experimentID})
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("Repo.GetByID: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[experimentRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Experiment{}, fmt.Errorf("Repo.GetByID: %w", domain.ErrNotFound)
		}

		return domain.Experiment{}, fmt.Errorf("Repo.GetByID: %w", err)
	}

	return row.toDomain()
}

func (r *Repo) Update(ctx context.Context, exp domain.Experiment) (domain.Experiment, error) {
	variantsJSON, err := marshalVariants(exp.Variants)
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("Repo.Update: %w", err)
	}

	const q = `
		UPDATE experiments
		SET name = @name, description = @description, status = @status,
		    traffic_percent = @traffic_percent, variants = @variants,
		    started_at = @started_at, ended_at = @ended_at, updated_at = NOW()
		WHERE id = @id AND project_id = @project_id
		RETURNING id, project_id, flag_id, name, description, status, traffic_percent, variants,
		          created_at, updated_at, started_at, ended_at`

	rows, err := r.db.Query(ctx, q, pgx.NamedArgs{
		"id":              exp.ID,
		"project_id":      exp.ProjectID,
		"name":            exp.Name,
		"description":     exp.Description,
		"status":          string(exp.Status),
		"traffic_percent": exp.TrafficPercent,
		"variants":        variantsJSON,
		"started_at":      exp.StartedAt,
		"ended_at":        exp.EndedAt,
	})
	if err != nil {
		return domain.Experiment{}, fmt.Errorf("Repo.Update: %w", err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[experimentRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Experiment{}, fmt.Errorf("Repo.Update: %w", domain.ErrNotFound)
		}

		return domain.Experiment{}, fmt.Errorf("Repo.Update: %w", err)
	}

	return row.toDomain()
}

func (r *Repo) Delete(ctx context.Context, projectID, experimentID uuid.UUID) error {
	const q = `DELETE FROM experiments WHERE project_id = @project_id AND id = @id`

	result, err := r.db.Exec(ctx, q, pgx.NamedArgs{"project_id": projectID, "id": experimentID})
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
