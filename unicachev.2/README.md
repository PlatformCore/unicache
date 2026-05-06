# unicache-enterprise

Enterprise shared cache library for Go microservices. ใช้เป็น shared lib/infra ได้ทันทีผ่าน `go get`.

```bash
go get github.com/diftappdev/unicache-enterprise
```

```go
store := cache.NewMemoryStore(cache.DefaultOptions())
c := cache.New(store, cache.Options{Namespace: "user-coupon", DefaultTTL: 5*time.Minute})
_ = c.Set(ctx, "u1", []byte("value"), time.Minute)
val, _ := c.Get(ctx, "u1")
```

จุดเด่น: TTL, sharded LRU, cleanup worker, namespace, singleflight กัน cache stampede, interfaces สำหรับ Redis/NATS/JetStream adapter, HTTP middleware, metrics hook, zero external dependency core.

> เปลี่ยน module path ใน `go.mod` ให้ตรง GitHub repo ของคุณก่อน push เช่น `github.com/diftappdev/unicache-enterprise`.


## Plug & Play cachex layer

This version includes `pkg/cachex`, a developer-friendly wrapper that lets services use cache without changing business logic deeply.

### Easiest usage

```go
c := cachex.MustFromConfig(cachex.Config{
    Namespace: "matching",
    Mode: cachex.ModeHybrid,
    DefaultTTL: 30 * time.Second,
})

user, err := cachex.Remember(ctx, c, "user:"+id, func(ctx context.Context) (User, error) {
    return repo.FindByID(ctx, id)
})
```

### Function decorator

```go
cachedFind := cachex.WrapFunc(c,
    func(id string) string { return "user:" + id },
    repo.FindByID,
    cachex.WithTTL(30*time.Second),
)
```

### Repository decorator

```go
cachedRepo := cachex.WrapRepository[string, User](c, userRepo, "user", 30*time.Second)
user, err := cachedRepo.FindByID(ctx, id)
```

### ENV config

```text
CACHE_ENABLED=true
CACHE_MODE=hybrid
CACHE_NAMESPACE=matching
CACHE_DEFAULT_TTL=30s
CACHE_MAX_ENTRIES=100000
```
