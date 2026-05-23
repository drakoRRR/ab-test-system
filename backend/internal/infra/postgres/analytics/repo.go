package analytics

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/analytics"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

const metricsQuery = `
SELECT
    variant_id,
    COUNT(*) FILTER (WHERE type = 'exposure')   AS exposures,
    COUNT(*) FILTER (WHERE type = 'conversion') AS conversions
FROM events
WHERE experiment_id = $1
GROUP BY variant_id`

func (r *Repo) GetMetrics(ctx context.Context, experimentID uuid.UUID) ([]domain.VariantMetrics, error) {
	rows, err := r.db.Query(ctx, metricsQuery, experimentID)
	if err != nil {
		return nil, fmt.Errorf("analytics.Repo.GetMetrics: %w", err)
	}
	defer rows.Close()

	var results []domain.VariantMetrics
	for rows.Next() {
		var m domain.VariantMetrics
		var vidBytes [16]byte
		if err := rows.Scan(&vidBytes, &m.Exposures, &m.Conversions); err != nil {
			return nil, fmt.Errorf("analytics.Repo.GetMetrics scan: %w", err)
		}
		m.VariantID = uuid.UUID(vidBytes)
		results = append(results, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("analytics.Repo.GetMetrics rows: %w", err)
	}

	return results, nil
}
