package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func Connect(ctx context.Context, redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URL: %w", err)
	}

	rdb := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("getting key %q: %w", key, err)
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshaling value for key %q: %w", key, err)
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("checking key %q: %w", key, err)
	}
	return n > 0, nil
}

// Incr atomically increments a counter key and sets its TTL on first creation.
// Subsequent calls extend the TTL only if the key is new (INCR returns 1).
func (c *Client) Incr(ctx context.Context, key string, ttl time.Duration) error {
	n, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("incrementing key %q: %w", key, err)
	}
	// Set TTL on first creation to bound counter lifetime.
	if n == 1 {
		if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
			return fmt.Errorf("setting TTL for key %q: %w", key, err)
		}
	}
	return nil
}
