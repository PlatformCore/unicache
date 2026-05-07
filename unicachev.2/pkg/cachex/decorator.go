package cachex

import (
	"context"
	"fmt"
	"time"
)

// Repository is a generic decorator for common Find/Get methods.
// For custom repositories, prefer WrapFunc because it does not require interfaces.
type Repository[ID comparable, Entity any] interface {
	FindByID(ctx context.Context, id ID) (Entity, error)
}

type CachedRepository[ID comparable, Entity any] struct {
	next   Repository[ID, Entity]
	cache  *Client
	prefix string
	ttl    time.Duration
}

func WrapRepository[ID comparable, Entity any](cache *Client, next Repository[ID, Entity], prefix string, ttl time.Duration) *CachedRepository[ID, Entity] {
	return &CachedRepository[ID, Entity]{next: next, cache: cache, prefix: prefix, ttl: ttl}
}

func (r *CachedRepository[ID, Entity]) FindByID(ctx context.Context, id ID) (Entity, error) {
	key := fmt.Sprintf("%s:%v", r.prefix, id)
	return Remember(ctx, r.cache, key, func(ctx context.Context) (Entity, error) {
		return r.next.FindByID(ctx, id)
	}, WithTTL(r.ttl))
}
