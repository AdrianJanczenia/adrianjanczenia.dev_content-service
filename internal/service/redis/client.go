package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisUniversalClient interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Close() error
}

type Client struct {
	client redisUniversalClient
}

func NewClient(redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opts)

	return &Client{client: rdb}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) SetToken(ctx context.Context, token string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, token, value, ttl).Err()
}

func (c *Client) ValidateAndDeleteToken(ctx context.Context, token string) (bool, error) {
	deletedCount, err := c.client.Del(ctx, token).Result()
	if err != nil {
		return false, err
	}

	return deletedCount > 0, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
