package cache

import (
	"context"
	"sync"
	"time"
)

type Client struct {
	store Store
	opt   Options
	mu    sync.Mutex
	calls map[string]*call
	pub   InvalidationPublisher
}
type call struct {
	wg  sync.WaitGroup
	val []byte
	ttl time.Duration
	err error
}

func New(store Store, opt Options) *Client {
	opt = opt.normalize()
	return &Client{store: store, opt: opt, calls: map[string]*call{}}
}
func (c *Client) WithInvalidationPublisher(p InvalidationPublisher) *Client { c.pub = p; return c }
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	e, err := c.store.Get(ctx, c.key(key))
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), e.Value...), nil
}
func (c *Client) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return c.store.Set(ctx, Entry{Key: c.key(key), Value: val}, ttl)
}
func (c *Client) Delete(ctx context.Context, key string) error {
	err := c.store.Delete(ctx, c.key(key))
	if err == nil && c.pub != nil {
		_ = c.pub.PublishInvalidation(ctx, c.key(key))
	}
	return err
}
func (c *Client) GetOrLoad(ctx context.Context, key string, loader Loader) ([]byte, error) {
	if v, err := c.Get(ctx, key); err == nil {
		return v, nil
	}
	if !c.opt.EnableSingleFlight {
		return c.load(ctx, key, loader)
	}
	ck := c.key(key)
	c.mu.Lock()
	if existing := c.calls[ck]; existing != nil {
		c.mu.Unlock()
		existing.wg.Wait()
		return append([]byte(nil), existing.val...), existing.err
	}
	ca := &call{}
	ca.wg.Add(1)
	c.calls[ck] = ca
	c.mu.Unlock()
	ca.val, ca.ttl, ca.err = loader(ctx, key)
	if ca.err == nil {
		ca.err = c.Set(ctx, key, ca.val, ca.ttl)
	}
	ca.wg.Done()
	c.mu.Lock()
	delete(c.calls, ck)
	c.mu.Unlock()
	return append([]byte(nil), ca.val...), ca.err
}
func (c *Client) load(ctx context.Context, key string, loader Loader) ([]byte, error) {
	v, ttl, err := loader(ctx, key)
	if err != nil {
		return nil, err
	}
	return v, c.Set(ctx, key, v, ttl)
}
func (c *Client) Close() error { return c.store.Close() }
func (c *Client) key(k string) string {
	if c.opt.Namespace == "" {
		return k
	}
	return c.opt.Namespace + ":" + k
}
