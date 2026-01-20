package task

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockRedis struct {
	getTokenFunc func(ctx context.Context, key string) (string, error)
	setTokenFunc func(ctx context.Context, token string, value interface{}, ttl time.Duration) error
	delTokenFunc func(ctx context.Context, key string) error
}

func (m *mockRedis) GetToken(ctx context.Context, key string) (string, error) {
	return m.getTokenFunc(ctx, key)
}
func (m *mockRedis) SetToken(ctx context.Context, token string, value interface{}, ttl time.Duration) error {
	return m.setTokenFunc(ctx, token, value, ttl)
}
func (m *mockRedis) DelToken(ctx context.Context, key string) error {
	return m.delTokenFunc(ctx, key)
}

func TestValidatePasswordTask_Execute(t *testing.T) {
	correctPass := "secret123"
	captchaID := "test-captcha"
	ctx := context.Background()

	t.Run("correct password", func(t *testing.T) {
		task := NewValidatePasswordTask(correctPass, &mockRedis{}, 10)
		err := task.Execute(ctx, "secret123", captchaID)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("incorrect password - decrement tries", func(t *testing.T) {
		c := captcha{Value: "XYZ", TriesLeft: 3, Solved: true}
		data, _ := json.Marshal(c)

		setCalled := false
		mock := &mockRedis{
			getTokenFunc: func(ctx context.Context, key string) (string, error) {
				return string(data), nil
			},
			setTokenFunc: func(ctx context.Context, token string, value interface{}, ttl time.Duration) error {
				setCalled = true
				var updated captcha
				json.Unmarshal([]byte(value.(string)), &updated)
				if updated.TriesLeft != 2 {
					t.Errorf("expected TriesLeft 2, got %d", updated.TriesLeft)
				}
				return nil
			},
		}

		task := NewValidatePasswordTask(correctPass, mock, 10)
		err := task.Execute(ctx, "wrong", captchaID)

		if err != errors.ErrInvalidPassword {
			t.Errorf("expected ErrInvalidPassword, got %v", err)
		}
		if !setCalled {
			t.Error("expected SetToken to be called")
		}
	})

	t.Run("incorrect password - delete captcha when no tries left", func(t *testing.T) {
		c := captcha{Value: "XYZ", TriesLeft: 1, Solved: true}
		data, _ := json.Marshal(c)

		delCalled := false
		mock := &mockRedis{
			getTokenFunc: func(ctx context.Context, key string) (string, error) {
				return string(data), nil
			},
			delTokenFunc: func(ctx context.Context, key string) error {
				delCalled = true
				return nil
			},
		}

		task := NewValidatePasswordTask(correctPass, mock, 10)
		err := task.Execute(ctx, "wrong", captchaID)

		if err != errors.ErrNoTriesLeft {
			t.Errorf("expected ErrNoTriesLeft, got %v", err)
		}
		if !delCalled {
			t.Error("expected DelToken to be called")
		}
	})

	t.Run("captcha not found", func(t *testing.T) {
		mock := &mockRedis{
			getTokenFunc: func(ctx context.Context, key string) (string, error) {
				return "", errors.ErrCaptchaNotFound
			},
		}

		task := NewValidatePasswordTask(correctPass, mock, 10)
		err := task.Execute(ctx, "wrong", captchaID)

		if err != errors.ErrCaptchaNotFound {
			t.Errorf("expected ErrCaptchaNotFound, got %v", err)
		}
	})
}
