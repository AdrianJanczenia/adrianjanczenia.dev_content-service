package download_cv

import (
	"errors"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockTokenValidator struct {
	validateFunc func(token string) (bool, error)
}

func (m *mockTokenValidator) ValidateAndDeleteToken(token string) (bool, error) {
	return m.validateFunc(token)
}

func TestProcess_DownloadCV(t *testing.T) {
	cvPaths := map[string]string{"pl": "/app/cv_pl.pdf"}

	tests := []struct {
		name         string
		token        string
		lang         string
		validateFunc func(string) (bool, error)
		wantPath     string
		wantErr      error
	}{
		{
			name:  "successful download",
			token: "valid",
			lang:  "pl",
			validateFunc: func(t string) (bool, error) {
				return true, nil
			},
			wantPath: "/app/cv_pl.pdf",
			wantErr:  nil,
		},
		{
			name:  "invalid token",
			token: "invalid",
			lang:  "pl",
			validateFunc: func(t string) (bool, error) {
				return false, nil
			},
			wantPath: "",
			wantErr:  appErrors.ErrCVExpired,
		},
		{
			name:  "unsupported language",
			token: "valid",
			lang:  "en",
			validateFunc: func(t string) (bool, error) {
				return true, nil
			},
			wantPath: "",
			wantErr:  appErrors.ErrUnsupportedLanguage,
		},
		{
			name:  "validator internal error",
			token: "valid",
			lang:  "pl",
			validateFunc: func(t string) (bool, error) {
				return false, errors.New("redis error")
			},
			wantPath: "",
			wantErr:  appErrors.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcess(&mockTokenValidator{validateFunc: tt.validateFunc}, cvPaths)
			path, err := p.Process(tt.token, tt.lang)

			if err != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if path != tt.wantPath {
				t.Errorf("Process() path = %v, wantPath %v", path, tt.wantPath)
			}
		})
	}
}
