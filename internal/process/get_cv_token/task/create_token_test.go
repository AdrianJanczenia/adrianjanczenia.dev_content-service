package task

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockTokenService struct {
	setTokenFunc func(ctx context.Context, token string, value interface{}, ttl time.Duration) error
}

func (m *mockTokenService) SetToken(ctx context.Context, token string, value interface{}, ttl time.Duration) error {
	return m.setTokenFunc(ctx, token, value, ttl)
}

func TestCreateTokenTask_Execute(t *testing.T) {
	tests := []struct {
		name         string
		setTokenFunc func(context.Context, string, interface{}, time.Duration) error
		wantErr      error
	}{
		{
			name: "success",
			setTokenFunc: func(ctx context.Context, token string, value interface{}, ttl time.Duration) error {
				if len(token) != 32 {
					return errors.New("invalid length")
				}
				match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", token)
				if !match {
					return errors.New("invalid charset")
				}
				return nil
			},
			wantErr: nil,
		},
		{
			name: "redis error",
			setTokenFunc: func(ctx context.Context, token string, value interface{}, ttl time.Duration) error {
				return errors.New("fail")
			},
			wantErr: appErrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockTokenService{setTokenFunc: tt.setTokenFunc}
			task := NewCreateTokenTask(m, time.Minute)
			_, err := task.Execute(context.Background())
			if err != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
