package task

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockDeleteCaptchaRedis struct {
	delTokenFunc func(ctx context.Context, key string) error
}

func (m *mockDeleteCaptchaRedis) DelToken(ctx context.Context, key string) error {
	return m.delTokenFunc(ctx, key)
}

func TestDeleteCaptchaTask_Execute(t *testing.T) {
	tests := []struct {
		name         string
		delTokenFunc func(context.Context, string) error
		wantErr      error
	}{
		{
			name: "success",
			delTokenFunc: func(ctx context.Context, key string) error {
				return nil
			},
			wantErr: nil,
		},
		{
			name: "error",
			delTokenFunc: func(ctx context.Context, key string) error {
				return errors.New("fail")
			},
			wantErr: appErrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockDeleteCaptchaRedis{delTokenFunc: tt.delTokenFunc}
			task := NewDeleteCaptchaTask(m)
			err := task.Execute(context.Background(), "id")
			if err != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
