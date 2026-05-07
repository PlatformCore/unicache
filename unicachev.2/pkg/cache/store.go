package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("cache: key not found")
	ErrClosed     = errors.New("cache: closed")
	ErrInvalidTTL = errors.New("cache: invalid ttl")
)

type Entry struct {
	Key       string
	Value     []byte
	ExpiresAt time.Time
	Version   uint64
	Metadata  map[string]string
}

func (e Entry) Expired(now time.Time) bool { return !e.ExpiresAt.IsZero() && now.After(e.ExpiresAt) }

type Store interface {
	Get(ctx context.Context, key string) (Entry, error)
	Set(ctx context.Context, entry Entry, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Has(ctx context.Context, key string) (bool, error)
	Close() error
}

type Loader func(ctx context.Context, key string) ([]byte, time.Duration, error)

type InvalidationPublisher interface {
	PublishInvalidation(ctx context.Context, key string) error
}
type InvalidationSubscriber interface {
	SubscribeInvalidation(ctx context.Context, fn func(context.Context, string)) error
}

type Metrics interface {
	IncHit(layer string)
	IncMiss(layer string)
	IncError(layer, op string)
	ObserveLatency(layer, op string, d time.Duration)
}

type NoopMetrics struct{}

func (NoopMetrics) IncHit(string)                                {}
func (NoopMetrics) IncMiss(string)                               {}
func (NoopMetrics) IncError(string, string)                      {}
func (NoopMetrics) ObserveLatency(string, string, time.Duration) {}
