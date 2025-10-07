package redis

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	client *redis.Client
}

func NewClient(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &Client{client: rdb}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) SetToken(token string, value interface{}, ttl time.Duration) error {
	return c.client.Set(context.Background(), token, value, ttl).Err()
}

func (c *Client) ValidateToken(token string) (bool, error) {
	err := c.client.Get(context.Background(), token).Err()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	deletedCount, err := c.client.Del(context.Background(), token).Result()
	if err != nil {
		log.Printf("WARN: failed to delete token %s after validation: %v", token, err)
	}
	if deletedCount == 0 {
		return false, errors.New("token already used")
	}

	return true, nil
}
