package flag

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	domain "github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/internal/domain/flag"
	"github.com/drakoRRR/ab-test-system/apps/ab-test-system/backend/pkg/cache"
)

const flagListTTL = 30 * time.Second

type CachedRepo struct {
	inner *Repo
	cache cache.Cache
	log   zerolog.Logger
}

func NewCachedRepo(inner *Repo, c cache.Cache, log zerolog.Logger) *CachedRepo {
	return &CachedRepo{inner: inner, cache: c, log: log}
}

func (r *CachedRepo) List(ctx context.Context, projectID uuid.UUID) ([]domain.Flag, error) {
	return cache.GetOrSet(
		ctx, r.cache, flagListKey(projectID), flagListTTL,
		func(err error) {
			r.log.Warn().Err(err).Str("key", flagListKey(projectID)).Msg("flag list cache error")
		},
		func() ([]domain.Flag, error) {
			return r.inner.List(ctx, projectID)
		},
	)
}

func (r *CachedRepo) Create(ctx context.Context, f domain.Flag) (domain.Flag, error) {
	created, err := r.inner.Create(ctx, f)
	if err != nil {
		return domain.Flag{}, err
	}

	if delErr := r.cache.Delete(ctx, flagListKey(f.ProjectID)); delErr != nil {
		r.log.Warn().Err(delErr).Str("key", flagListKey(f.ProjectID)).Msg("flag list cache invalidation failed")
	}

	return created, nil
}

func (r *CachedRepo) GetByKey(ctx context.Context, projectID uuid.UUID, key string) (domain.Flag, error) {
	return r.inner.GetByKey(ctx, projectID, key)
}

func (r *CachedRepo) Update(ctx context.Context, f domain.Flag) (domain.Flag, error) {
	updated, err := r.inner.Update(ctx, f)
	if err != nil {
		return domain.Flag{}, err
	}

	if delErr := r.cache.Delete(ctx, flagListKey(f.ProjectID)); delErr != nil {
		r.log.Warn().Err(delErr).Str("key", flagListKey(f.ProjectID)).Msg("flag list cache invalidation failed")
	}

	return updated, nil
}

func (r *CachedRepo) Delete(ctx context.Context, projectID uuid.UUID, key string) error {
	if err := r.inner.Delete(ctx, projectID, key); err != nil {
		return err
	}

	if delErr := r.cache.Delete(ctx, flagListKey(projectID)); delErr != nil {
		r.log.Warn().Err(delErr).Str("key", flagListKey(projectID)).Msg("flag list cache invalidation failed")
	}

	return nil
}

func flagListKey(projectID uuid.UUID) string {
	return fmt.Sprintf("flags:%s", projectID)
}
