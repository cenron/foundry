package cache_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cenron/foundry/internal/cache"
)

func testRedisURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}
	return url
}

func TestGetSet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := cache.Connect(ctx, testRedisURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = c.Close() }()

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	key := "test:cache:getset"
	want := payload{Name: "foundry", Count: 42}

	if err := c.Set(ctx, key, want, 10*time.Second); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	var got payload
	if err := c.Get(ctx, key, &got); err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}

	// Cleanup
	_ = c.Delete(ctx, key)
}

func TestDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := cache.Connect(ctx, testRedisURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = c.Close() }()

	key := "test:cache:delete"
	_ = c.Set(ctx, key, "hello", 10*time.Second)

	if err := c.Delete(ctx, key); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	exists, err := c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists() error: %v", err)
	}
	if exists {
		t.Error("key should not exist after delete")
	}
}

func TestExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c, err := cache.Connect(ctx, testRedisURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = c.Close() }()

	key := "test:cache:exists"

	exists, _ := c.Exists(ctx, key)
	if exists {
		t.Error("key should not exist initially")
	}

	_ = c.Set(ctx, key, "data", 10*time.Second)

	exists, err = c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists() error: %v", err)
	}
	if !exists {
		t.Error("key should exist after set")
	}

	_ = c.Delete(ctx, key)
}
