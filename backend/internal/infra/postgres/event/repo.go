package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainevent "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/event"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

type eventRow struct {
	ID           uuid.UUID `db:"id"`
	ProjectID    uuid.UUID `db:"project_id"`
	UserID       string    `db:"user_id"`
	ExperimentID uuid.UUID `db:"experiment_id"`
	VariantID    uuid.UUID `db:"variant_id"`
	Type         string    `db:"type"`
	Name         string    `db:"name"`
	Value        float64   `db:"value"`
	Timestamp    time.Time `db:"ts"`
}

func fromDomain(e domainevent.Event) eventRow {
	return eventRow{
		ID:           e.ID,
		ProjectID:    e.ProjectID,
		UserID:       e.UserID,
		ExperimentID: e.ExperimentID,
		VariantID:    e.VariantID,
		Type:         string(e.Type),
		Name:         e.Name,
		Value:        e.Value,
		Timestamp:    e.Timestamp,
	}
}

func (r *Repo) BulkInsert(ctx context.Context, events []domainevent.Event) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	const q = `
		INSERT INTO events (id, project_id, user_id, experiment_id, variant_id, type, name, value, ts)
		VALUES (@id, @project_id, @user_id, @experiment_id, @variant_id, @type, @name, @value, @ts)
		ON CONFLICT (id) DO NOTHING`

	for _, e := range events {
		row := fromDomain(e)
		batch.Queue(q, pgx.NamedArgs{
			"id":            row.ID,
			"project_id":    row.ProjectID,
			"user_id":       row.UserID,
			"experiment_id": row.ExperimentID,
			"variant_id":    row.VariantID,
			"type":          row.Type,
			"name":          row.Name,
			"value":         row.Value,
			"ts":            row.Timestamp,
		})
	}

	results := r.db.SendBatch(ctx, batch)

	var execErr error
	for range events {
		if _, err := results.Exec(); err != nil {
			execErr = fmt.Errorf("event.Repo.BulkInsert: %w", err)
			break
		}
	}

	if closeErr := results.Close(); closeErr != nil && execErr == nil {
		return fmt.Errorf("event.Repo.BulkInsert: close batch: %w", closeErr)
	}
	return execErr
}
