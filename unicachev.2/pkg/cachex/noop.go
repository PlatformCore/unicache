package cachex

import (
	"context"
	"time"

	"github.com/diftappdev/unicache-enterprise/pkg/cache"
)

type NoopCache struct{}

func (NoopCache) Get(ctx context.Context, key string) ([]byte, error) { return nil, cache.ErrNotFound }
func (NoopCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return nil
}
func (NoopCache) Delete(ctx context.Context, key string) error { return nil }
func (NoopCache) Close() error                                 { return nil }
func (NoopCache) GetOrLoad(ctx context.Context, key string, loader cache.Loader) ([]byte, error) {
	v, _, err := loader(ctx, key)
	return v, err
}
