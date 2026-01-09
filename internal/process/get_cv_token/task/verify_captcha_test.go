package task

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockVerifyCaptchaRedis struct {
	getTokenFunc func(ctx context.Context, key string) (string, error)
}

func (m *mockVerifyCaptchaRedis) GetToken(ctx context.Context, key string) (string, error) {
	return m.getTokenFunc(ctx, key)
}

func TestVerifyCaptchaTask_Execute(t *testing.T) {
	tests := []struct {
		name         string
		getTokenFunc func(context.Context, string) (string, error)
		wantErr      error
	}{
		{
			name: "success",
			getTokenFunc: func(ctx context.Context, key string) (string, error) {
				return `{"solved": true}`, nil
			},
			wantErr: nil,
		},
		{
			name: "not found",
			getTokenFunc: func(ctx context.Context, key string) (string, error) {
				return "", errors.New("redis error")
			},
			wantErr: appErrors.ErrCaptchaNotFound,
		},
		{
			name: "not solved",
			getTokenFunc: func(ctx context.Context, key string) (string, error) {
				return `{"solved": false}`, nil
			},
			wantErr: appErrors.ErrCaptchaNotSolved,
		},
		{
			name: "invalid json",
			getTokenFunc: func(ctx context.Context, key string) (string, error) {
				return `invalid`, nil
			},
			wantErr: appErrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockVerifyCaptchaRedis{getTokenFunc: tt.getTokenFunc}
			task := NewVerifyCaptchaTask(m)
			err := task.Execute(context.Background(), "id")
			if err != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
