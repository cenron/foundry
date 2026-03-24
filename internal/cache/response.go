package cache

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	responseCachePrefix = "response:"
	hitCountPrefix      = "response_hits:"
	missCountPrefix     = "response_misses:"
	defaultResponseTTL  = 5 * time.Minute
)

// ResponseCacheKey identifies a cacheable prompt/response pair.
type ResponseCacheKey struct {
	ProjectID string
	Prompt    string
	Model     string
}

// Hash returns a deterministic cache key string.
func (k ResponseCacheKey) Hash() string {
	raw := fmt.Sprintf("%s:%s:%s", k.ProjectID, k.Model, k.Prompt)
	sum := sha256.Sum256([]byte(raw))
	return responseCachePrefix + fmt.Sprintf("%x", sum[:16])
}

type ResponseCache struct {
	client *Client
}

func NewResponseCache(client *Client) *ResponseCache {
	return &ResponseCache{client: client}
}

// Get retrieves a cached response. Returns the response and whether it was a hit.
func (c *ResponseCache) Get(ctx context.Context, key ResponseCacheKey) (string, bool, error) {
	hash := key.Hash()

	var response string
	err := c.client.Get(ctx, hash, &response)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.incrementCounter(ctx, missCountPrefix+key.ProjectID)
			return "", false, nil
		}
		return "", false, fmt.Errorf("getting cached response: %w", err)
	}

	c.incrementCounter(ctx, hitCountPrefix+key.ProjectID)
	return response, true, nil
}

// Set stores a response in the cache with a TTL.
func (c *ResponseCache) Set(ctx context.Context, key ResponseCacheKey, response string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = defaultResponseTTL
	}
	return c.client.Set(ctx, key.Hash(), response, ttl)
}

// HitRate returns the cache hit rate for a project (0.0–1.0).
func (c *ResponseCache) HitRate(ctx context.Context, projectID string) (float64, error) {
	var hits, misses int

	if err := c.client.Get(ctx, hitCountPrefix+projectID, &hits); err != nil {
		hits = 0
	}
	if err := c.client.Get(ctx, missCountPrefix+projectID, &misses); err != nil {
		misses = 0
	}

	total := hits + misses
	if total == 0 {
		return 0, nil
	}

	return float64(hits) / float64(total), nil
}

func (c *ResponseCache) incrementCounter(ctx context.Context, key string) {
	if err := c.client.Incr(ctx, key, 24*time.Hour); err != nil {
		log.Printf("response cache: incrementing counter %q: %v", key, err)
	}
}
