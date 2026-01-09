package redis

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestClient_GetToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{client: db}
	ctx := context.Background()
	key := "test-key"
	val := "test-val"

	t.Run("success", func(t *testing.T) {
		mock.ExpectGet(key).SetVal(val)
		got, err := client.GetToken(ctx, key)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got != val {
			t.Errorf("got %s, want %s", got, val)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectGet(key).SetErr(errors.New("not found"))
		_, err := client.GetToken(ctx, key)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestClient_DelToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{client: db}
	ctx := context.Background()
	key := "test-key"

	t.Run("success", func(t *testing.T) {
		mock.ExpectDel(key).SetVal(1)
		err := client.DelToken(ctx, key)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectDel(key).SetErr(errors.New("del error"))
		err := client.DelToken(ctx, key)
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
		mock.ExpectDel(token).SetVal(0)

		valid, err := client.ValidateAndDeleteToken(ctx, token)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if valid {
			t.Error("expected valid to be false")
		}
	})

	t.Run("delete error", func(t *testing.T) {
		mock.ExpectDel(token).SetErr(errors.New("delete error"))

		valid, err := client.ValidateAndDeleteToken(ctx, token)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if valid {
			t.Error("expected valid to be false")
		}
	})
}
