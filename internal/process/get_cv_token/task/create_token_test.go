package task

import (
	"errors"
	"testing"
	"time"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockTokenService struct {
	setTokenFunc func(token string, value interface{}, ttl time.Duration) error
}

func (m *mockTokenService) SetToken(token string, value interface{}, ttl time.Duration) error {
	return m.setTokenFunc(token, value, ttl)
}

func TestCreateTokenTask_Execute(t *testing.T) {
	tests := []struct {
		name         string
		setTokenFunc func(token string, value interface{}, ttl time.Duration) error
		wantErr      error
	}{
		{
			name: "successful token creation",
			setTokenFunc: func(token string, value interface{}, ttl time.Duration) error {
				return nil
			},
			wantErr: nil,
		},
		{
			name: "token service error",
			setTokenFunc: func(token string, value interface{}, ttl time.Duration) error {
				return errors.New("redis down")
			},
			wantErr: appErrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTokenService{setTokenFunc: tt.setTokenFunc}
			task := NewCreateTokenTask(mock, time.Minute)

			token, err := task.Execute()

			if err != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && token == "" {
				t.Error("Execute() expected token but got empty string")
			}
		})
	}
}
