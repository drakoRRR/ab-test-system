package experiment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/experiment"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/pkg/cache"
)

const experimentListTTL = 30 * time.Second

type CachedRepo struct {
	inner *Repo
	cache cache.Cache
	log   zerolog.Logger
}

func NewCachedRepo(inner *Repo, c cache.Cache, log zerolog.Logger) *CachedRepo {
	return &CachedRepo{inner: inner, cache: c, log: log}
}

func (r *CachedRepo) List(ctx context.Context, projectID uuid.UUID) ([]domain.Experiment, error) {
	return cache.GetOrSet(
		ctx, r.cache, experimentListKey(projectID), experimentListTTL,
		func(err error) {
			r.log.Warn().Err(err).Str("key", experimentListKey(projectID)).Msg("experiment list cache error")
		},
		func() ([]domain.Experiment, error) {
			return r.inner.List(ctx, projectID)
		},
	)
}

func (r *CachedRepo) Create(ctx context.Context, exp domain.Experiment) (domain.Experiment, error) {
	created, err := r.inner.Create(ctx, exp)
	if err != nil {
		return domain.Experiment{}, err
	}

	if delErr := r.cache.Delete(ctx, experimentListKey(exp.ProjectID)); delErr != nil {
		r.log.Warn().
			Err(delErr).
			Str("key", experimentListKey(exp.ProjectID)).
			Msg("experiment list cache invalidation failed")
	}

	return created, nil
}

func (r *CachedRepo) GetByID(ctx context.Context, projectID, experimentID uuid.UUID) (domain.Experiment, error) {
	return r.inner.GetByID(ctx, projectID, experimentID)
}

func (r *CachedRepo) Update(ctx context.Context, exp domain.Experiment) (domain.Experiment, error) {
	updated, err := r.inner.Update(ctx, exp)
	if err != nil {
		return domain.Experiment{}, err
	}

	if delErr := r.cache.Delete(ctx, experimentListKey(exp.ProjectID)); delErr != nil {
		r.log.Warn().
			Err(delErr).
			Str("key", experimentListKey(exp.ProjectID)).
			Msg("experiment list cache invalidation failed")
	}

	return updated, nil
}

func (r *CachedRepo) Delete(ctx context.Context, projectID, experimentID uuid.UUID) error {
	if err := r.inner.Delete(ctx, projectID, experimentID); err != nil {
		return err
	}

	if delErr := r.cache.Delete(ctx, experimentListKey(projectID)); delErr != nil {
		r.log.Warn().
			Err(delErr).
			Str("key", experimentListKey(projectID)).
			Msg("experiment list cache invalidation failed")
	}

	return nil
}

func experimentListKey(projectID uuid.UUID) string {
	return fmt.Sprintf("experiments:%s", projectID)
}
