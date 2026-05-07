package cachex

import (
	"context"
	"log/slog"
	"time"

	"github.com/diftappdev/unicache-enterprise/pkg/cache"
)

type Middleware func(cache.Loader) cache.Loader

func LoaderTimeout(timeout time.Duration) Middleware {
	return func(next cache.Loader) cache.Loader {
		return func(ctx context.Context, key string) ([]byte, time.Duration, error) {
			if timeout <= 0 {
				return next(ctx, key)
			}
			cctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return next(cctx, key)
		}
	}
}

func LoaderLogging(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next cache.Loader) cache.Loader {
		return func(ctx context.Context, key string) ([]byte, time.Duration, error) {
			st := time.Now()
			b, ttl, err := next(ctx, key)
			logger.Debug("cache loader", "key", key, "ttl", ttl, "latency_ms", time.Since(st).Milliseconds(), "err", err)
			return b, ttl, err
		}
	}
}

func LoaderRetry(times int, delay time.Duration) Middleware {
	return func(next cache.Loader) cache.Loader {
		return func(ctx context.Context, key string) ([]byte, time.Duration, error) {
			if times <= 0 {
				return next(ctx, key)
			}
			var last error
			for i := 0; i <= times; i++ {
				b, ttl, err := next(ctx, key)
				if err == nil {
					return b, ttl, nil
				}
				last = err
				if delay > 0 {
					select {
					case <-ctx.Done():
						return nil, 0, ctx.Err()
					case <-time.After(delay):
					}
				}
			}
			return nil, 0, last
		}
	}
}
