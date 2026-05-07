package cachex

import "time"

type Option func(*Client)

func WithConfig(cfg Config) Option { return func(c *Client) { c.cfg = cfg.Normalize() } }
func WithCodec(codec Codec) Option {
	return func(c *Client) {
		if codec != nil {
			c.codec = codec
		}
	}
}
func WithKeyPrefix(prefix string) Option     { return func(c *Client) { c.prefix = prefix } }
func WithMiddleware(mw ...Middleware) Option { return func(c *Client) { c.mw = append(c.mw, mw...) } }

type CallOptions struct {
	TTL         time.Duration
	NegativeTTL time.Duration
	KeyPrefix   string
}
type CallOption func(*CallOptions)

func defaultCallOptions(cfg Config) CallOptions {
	cfg = cfg.Normalize()
	return CallOptions{TTL: cfg.DefaultTTL, NegativeTTL: cfg.NegativeTTL}
}
func WithTTL(ttl time.Duration) CallOption { return func(o *CallOptions) { o.TTL = ttl } }
func WithNegativeTTL(ttl time.Duration) CallOption {
	return func(o *CallOptions) { o.NegativeTTL = ttl }
}
func WithCallKeyPrefix(prefix string) CallOption {
	return func(o *CallOptions) { o.KeyPrefix = prefix }
}
