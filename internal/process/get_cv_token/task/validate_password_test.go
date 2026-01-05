package task

import (
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

func TestValidatePasswordTask_Execute(t *testing.T) {
	correctPass := "secret123"
	task := NewValidatePasswordTask(correctPass)

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "correct password",
			password: "secret123",
			wantErr:  nil,
		},
		{
			name:     "incorrect password",
			password: "wrong",
			wantErr:  errors.ErrInvalidPassword,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  errors.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := task.Execute(tt.password)
			if err != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
