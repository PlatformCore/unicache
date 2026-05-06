package main

import (
	"context"
	"fmt"
	"time"

	"github.com/diftappdev/unicache-enterprise/pkg/cachex"
)

type User struct {
	ID   string
	Name string
}

func main() {
	ctx := context.Background()
	c := cachex.MustFromConfig(cachex.Config{Namespace: "demo", Mode: cachex.ModeHybrid, DefaultTTL: time.Minute})
	defer c.Close()

	user, err := cachex.Remember(ctx, c, "user:42", func(ctx context.Context) (User, error) {
		return User{ID: "42", Name: "Dift"}, nil
	}, cachex.WithTTL(30*time.Second))
	if err != nil {
		panic(err)
	}
	fmt.Println(user.Name)

	loadPrice := func(ctx context.Context, productID string) (int, error) { return 199, nil }
	cachedPrice := cachex.WrapFunc(c, func(productID string) string { return "price:" + productID }, loadPrice, cachex.WithTTL(10*time.Second))
	price, _ := cachedPrice(ctx, "p1")
	fmt.Println(price)
}
