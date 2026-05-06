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
