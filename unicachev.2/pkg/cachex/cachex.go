package cachex

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/diftappdev/unicache-enterprise/pkg/cache"
)

var ErrDisabled = errors.New("cachex: disabled")

// BytesCache is the minimum contract supported by unicache Client and hybrid Cache.
type BytesCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GetOrLoad(ctx context.Context, key string, loader cache.Loader) ([]byte, error)
	Close() error
}

type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(b []byte, v any) error
}

type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error)   { return json.Marshal(v) }
func (JSONCodec) Unmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

// Client is a plug-and-play wrapper over unicache/hybridcache.
type Client struct {
	raw    BytesCache
	cfg    Config
	codec  Codec
	prefix string
	mw     []Middleware
}

func New(raw BytesCache, opts ...Option) *Client {
	c := &Client{raw: raw, cfg: DefaultConfig(), codec: JSONCodec{}}
	for _, opt := range opts {
		opt(c)
	}
	c.cfg = c.cfg.Normalize()
	return c
}

func NewFromConfig(cfg Config, opts ...Option) (*Client, error) {
	cfg = cfg.Normalize()
	if !cfg.Enabled {
		return New(NoopCache{}, append(opts, WithConfig(cfg))...), nil
	}
	raw, err := buildRaw(cfg)
	if err != nil {
		return nil, err
	}
	return New(raw, append(opts, WithConfig(cfg))...), nil
}

func MustFromConfig(cfg Config, opts ...Option) *Client {
	c, err := NewFromConfig(cfg, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func FromEnv(opts ...Option) (*Client, error) { return NewFromConfig(ConfigFromEnv(), opts...) }
func MustFromEnv(opts ...Option) *Client      { return MustFromConfig(ConfigFromEnv(), opts...) }

func buildRaw(cfg Config) (BytesCache, error) {
	opt := cache.Options{
		Namespace: cfg.Namespace, DefaultTTL: cfg.DefaultTTL, MaxEntries: cfg.MaxEntries,
		MaxBytes: cfg.MaxBytes, Shards: cfg.Shards, CleanupInterval: cfg.CleanupInterval,
		StaleWhileRevalidate: cfg.StaleTTL, NegativeTTL: cfg.NegativeTTL,
		EnableSingleFlight: cfg.EnableSingleFlight,
	}

	// unicache has one built-in layer. ModeHybrid falls back to the same memory client
	// so services can share the same config across unicache/hybridcache without crashing.
	return cache.New(cache.NewMemoryStore(opt), opt), nil
}

func (c *Client) Raw() BytesCache { return c.raw }
func (c *Client) Close() error {
	if c.raw == nil {
		return nil
	}
	return c.raw.Close()
}
func (c *Client) key(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + ":" + key
}

func (c *Client) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return c.raw.Get(ctx, c.key(key))
}
func (c *Client) SetBytes(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.cfg.DefaultTTL
	}
	return c.raw.Set(ctx, c.key(key), val, ttl)
}
func (c *Client) Invalidate(ctx context.Context, key string) error {
	return c.raw.Delete(ctx, c.key(key))
}

func Get[T any](ctx context.Context, c *Client, key string) (T, error) {
	var out T
	b, err := c.GetBytes(ctx, key)
	if err != nil {
		return out, err
	}
	if err := c.codec.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}

func Set[T any](ctx context.Context, c *Client, key string, val T, ttl time.Duration) error {
	b, err := c.codec.Marshal(val)
	if err != nil {
		return err
	}
	return c.SetBytes(ctx, key, b, ttl)
}

// Remember is the easiest API: cache miss -> call loader -> marshal -> set cache -> return value.
func Remember[T any](ctx context.Context, c *Client, key string, loader func(context.Context) (T, error), opts ...CallOption) (T, error) {
	var zero T
	call := defaultCallOptions(c.cfg)
	for _, opt := range opts {
		opt(&call)
	}
	if call.KeyPrefix != "" {
		key = call.KeyPrefix + ":" + key
	}

	finalLoader := func(ctx context.Context, _ string) ([]byte, time.Duration, error) {
		v, err := loader(ctx)
		if err != nil {
			return nil, call.NegativeTTL, err
		}
		b, err := c.codec.Marshal(v)
		if err != nil {
			return nil, 0, err
		}
		return b, call.TTL, nil
	}
	for i := len(c.mw) - 1; i >= 0; i-- {
		finalLoader = c.mw[i](finalLoader)
	}

	b, err := c.raw.GetOrLoad(ctx, c.key(key), finalLoader)
	if err != nil {
		return zero, err
	}
	var out T
	if err := c.codec.Unmarshal(b, &out); err != nil {
		return zero, err
	}
	return out, nil
}

// WrapFunc converts any function into a cached function without changing business logic.
func WrapFunc[Req any, Res any](c *Client, keyFn func(Req) string, fn func(context.Context, Req) (Res, error), opts ...CallOption) func(context.Context, Req) (Res, error) {
	return func(ctx context.Context, req Req) (Res, error) {
		return Remember(ctx, c, keyFn(req), func(ctx context.Context) (Res, error) { return fn(ctx, req) }, opts...)
	}
}
