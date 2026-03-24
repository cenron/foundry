package cache_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cenron/foundry/internal/cache"
)

func setupResponseCache(t *testing.T) *cache.ResponseCache {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}

	client, err := cache.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	t.Cleanup(func() { _ = client.Close() })

	return cache.NewResponseCache(client)
}

func TestResponseCache_SetAndGet(t *testing.T) {
	rc := setupResponseCache(t)
	ctx := context.Background()

	key := cache.ResponseCacheKey{
		ProjectID: "proj-test-1",
		Prompt:    "Write a hello world function",
		Model:     "sonnet",
	}

	// Set
	err := rc.Set(ctx, key, "def hello(): print('hello')", 10*time.Second)
	if err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Get — should be a hit
	response, hit, err := rc.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if !hit {
		t.Fatal("expected cache hit")
	}
	if response != "def hello(): print('hello')" {
		t.Errorf("response = %q", response)
	}
}

func TestResponseCache_Miss(t *testing.T) {
	rc := setupResponseCache(t)
	ctx := context.Background()

	key := cache.ResponseCacheKey{
		ProjectID: "proj-test-miss",
		Prompt:    "nonexistent prompt",
		Model:     "haiku",
	}

	_, hit, err := rc.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if hit {
		t.Fatal("expected cache miss")
	}
}

func TestResponseCache_DifferentModels(t *testing.T) {
	rc := setupResponseCache(t)
	ctx := context.Background()

	baseKey := cache.ResponseCacheKey{
		ProjectID: "proj-test-model",
		Prompt:    "same prompt",
	}

	// Set for sonnet
	sonnetKey := baseKey
	sonnetKey.Model = "sonnet"
	_ = rc.Set(ctx, sonnetKey, "sonnet response", 10*time.Second)

	// Set for haiku
	haikuKey := baseKey
	haikuKey.Model = "haiku"
	_ = rc.Set(ctx, haikuKey, "haiku response", 10*time.Second)

	// Get sonnet
	resp, hit, _ := rc.Get(ctx, sonnetKey)
	if !hit || resp != "sonnet response" {
		t.Errorf("sonnet: hit=%v, resp=%q", hit, resp)
	}

	// Get haiku
	resp, hit, _ = rc.Get(ctx, haikuKey)
	if !hit || resp != "haiku response" {
		t.Errorf("haiku: hit=%v, resp=%q", hit, resp)
	}
}

func TestResponseCache_HitRate(t *testing.T) {
	rc := setupResponseCache(t)
	ctx := context.Background()

	projectID := "proj-test-hitrate"
	key := cache.ResponseCacheKey{
		ProjectID: projectID,
		Prompt:    "cached prompt",
		Model:     "sonnet",
	}

	// Set a value
	_ = rc.Set(ctx, key, "cached", 10*time.Second)

	// 2 hits
	_, _, _ = rc.Get(ctx, key)
	_, _, _ = rc.Get(ctx, key)

	// 1 miss
	missKey := cache.ResponseCacheKey{ProjectID: projectID, Prompt: "miss", Model: "sonnet"}
	_, _, _ = rc.Get(ctx, missKey)

	rate, err := rc.HitRate(ctx, projectID)
	if err != nil {
		t.Fatalf("HitRate() error: %v", err)
	}

	// 2 hits / 3 total ≈ 0.667
	if rate < 0.6 || rate > 0.7 {
		t.Errorf("hit rate = %f, want ~0.667", rate)
	}
}
