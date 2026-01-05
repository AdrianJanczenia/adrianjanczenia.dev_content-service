package get_cv_token

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockValidatePasswordTask struct {
	executeFunc func(password string) error
}

func (m *mockValidatePasswordTask) Execute(password string) error {
	return m.executeFunc(password)
}

type mockCreateTokenTask struct {
	executeFunc func(ctx context.Context) (string, error)
}

func (m *mockCreateTokenTask) Execute(ctx context.Context) (string, error) {
	return m.executeFunc(ctx)
}

func TestProcess_GetCVToken(t *testing.T) {
	cvPaths := map[string]string{"pl": "/path/pl.pdf", "en": "/path/en.pdf"}

	tests := []struct {
		name            string
		password        string
		lang            string
		validateFunc    func(string) error
		createTokenFunc func(context.Context) (string, error)
		wantErr         error
		wantToken       string
	}{
		{
			name:            "successful process",
			password:        "pass",
			lang:            "pl",
			validateFunc:    func(p string) error { return nil },
			createTokenFunc: func(context.Context) (string, error) { return "valid-token", nil },
			wantErr:         nil,
			wantToken:       "valid-token",
		},
		{
			name:            "unsupported language",
			password:        "pass",
			lang:            "de",
			validateFunc:    func(p string) error { return nil },
			createTokenFunc: func(context.Context) (string, error) { return "", nil },
			wantErr:         appErrors.ErrUnsupportedLanguage,
			wantToken:       "",
		},
		{
			name:            "invalid password",
			password:        "wrong",
			lang:            "en",
			validateFunc:    func(p string) error { return appErrors.ErrInvalidPassword },
			createTokenFunc: func(context.Context) (string, error) { return "", nil },
			wantErr:         appErrors.ErrInvalidPassword,
			wantToken:       "",
		},
		{
			name:            "token creation error",
			password:        "pass",
			lang:            "pl",
			validateFunc:    func(p string) error { return nil },
			createTokenFunc: func(context.Context) (string, error) { return "", errors.New("fail") },
			wantErr:         errors.New("fail"),
			wantToken:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcess(
				&mockValidatePasswordTask{executeFunc: tt.validateFunc},
				&mockCreateTokenTask{executeFunc: tt.createTokenFunc},
				cvPaths,
			)

			token, err := p.Process(context.Background(), tt.password, tt.lang)

			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if token != tt.wantToken {
				t.Errorf("Process() token = %v, wantToken %v", token, tt.wantToken)
			}
		})
	}
}
