package main

import (
	"context"
	"fmt"
	"github.com/diftappdev/unicache-enterprise/pkg/cache"
	"time"
)

func main() {
	ctx := context.Background()
	c := cache.New(cache.NewMemoryStore(cache.DefaultOptions()), cache.Options{Namespace: "example"})
	defer c.Close()
	_ = c.Set(ctx, "hello", []byte("world"), time.Minute)
	v, _ := c.Get(ctx, "hello")
	fmt.Println(string(v))
}
