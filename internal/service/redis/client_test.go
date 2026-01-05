package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

func TestClient_Ping(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{client: db}
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectPing().SetVal("PONG")
		err := client.Ping(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectPing().SetErr(errors.New("connection failed"))
		err := client.Ping(ctx)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestClient_SetToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{client: db}
	ctx := context.Background()
	token := "test-token"
	ttl := time.Minute

	t.Run("success", func(t *testing.T) {
		mock.ExpectSet(token, "valid", ttl).SetVal("OK")
		err := client.SetToken(ctx, token, "valid", ttl)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectSet(token, "valid", ttl).SetErr(errors.New("redis error"))
		err := client.SetToken(ctx, token, "valid", ttl)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestClient_ValidateAndDeleteToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{client: db}
	ctx := context.Background()
	token := "test-token"

	t.Run("success", func(t *testing.T) {
		mock.ExpectGet(token).SetVal("valid")
		mock.ExpectDel(token).SetVal(1)

		valid, err := client.ValidateAndDeleteToken(ctx, token)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !valid {
			t.Error("expected valid to be true")
		}
	})

	t.Run("token not found", func(t *testing.T) {
		mock.ExpectGet(token).SetErr(redis.Nil)

		valid, err := client.ValidateAndDeleteToken(ctx, token)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if valid {
			t.Error("expected valid to be false")
		}
	})

	t.Run("get error", func(t *testing.T) {
		mock.ExpectGet(token).SetErr(errors.New("db error"))

		valid, err := client.ValidateAndDeleteToken(ctx, token)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if valid {
			t.Error("expected valid to be false")
		}
	})

	t.Run("delete error", func(t *testing.T) {
		mock.ExpectGet(token).SetVal("valid")
		mock.ExpectDel(token).SetErr(errors.New("delete error"))

		valid, err := client.ValidateAndDeleteToken(ctx, token)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if valid {
			t.Error("expected valid to be false because delete failed")
		}
	})
}
