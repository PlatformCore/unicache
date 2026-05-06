package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStoreTTL(t *testing.T) {
	ctx := context.Background()
	c := New(NewMemoryStore(Options{DefaultTTL: 10 * time.Millisecond}), Options{})
	if err := c.Set(ctx, "k", []byte("v"), 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if _, err := c.Get(ctx, "k"); err != ErrNotFound {
		t.Fatalf("want ErrNotFound got %v", err)
	}
}
func BenchmarkSetGet(b *testing.B) {
	ctx := context.Background()
	c := New(NewMemoryStore(DefaultOptions()), Options{})
	for i := 0; i < b.N; i++ {
		_ = c.Set(ctx, "k", []byte("v"), time.Minute)
		_, _ = c.Get(ctx, "k")
	}
}
