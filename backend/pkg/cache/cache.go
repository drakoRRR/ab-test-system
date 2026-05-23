package cache

import (
	"context"
	"encoding/json"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// GetOrSet returns the cached value for key if present, otherwise calls loader,
// stores the result, and returns it. Cache errors never fail the operation —
// onErr is called so the caller can log them; execution always continues.
func GetOrSet[T any](
	ctx context.Context,
	c Cache,
	key string,
	ttl time.Duration,
	onErr func(error),
	loader func() (T, error),
) (T, error) {
	raw, ok, err := c.Get(ctx, key)
	if err != nil {
		onErr(err)
	} else if ok {
		var val T
		if jsonErr := json.Unmarshal(raw, &val); jsonErr == nil {
			return val, nil
		}
	}

	val, err := loader()
	if err != nil {
		return val, err
	}

	raw, jsonErr := json.Marshal(val)
	if jsonErr != nil {
		onErr(jsonErr)
		return val, nil
	}

	if setErr := c.Set(ctx, key, raw, ttl); setErr != nil {
		onErr(setErr)
	}

	return val, nil
}
