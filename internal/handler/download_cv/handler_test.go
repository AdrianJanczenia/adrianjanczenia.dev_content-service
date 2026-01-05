package download_cv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockDownloadCVProcess struct {
	processFunc func(ctx context.Context, token, lang string) (string, error)
}

func (m *mockDownloadCVProcess) Process(ctx context.Context, token, lang string) (string, error) {
	return m.processFunc(ctx, token, lang)
}

func TestHandler_DownloadCV(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		processFunc func(context.Context, string, string) (string, error)
		wantStatus  int
	}{
		{
			name:       "wrong method",
			method:     http.MethodPost,
			url:        "/download/cv",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "missing token",
			method:     http.MethodGet,
			url:        "/download/cv?lang=pl",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing lang",
			method:     http.MethodGet,
			url:        "/download/cv?token=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "process error",
			method: http.MethodGet,
			url:    "/download/cv?token=abc&lang=pl",
			processFunc: func(ctx context.Context, t, l string) (string, error) {
				return "", errors.ErrCVExpired
			},
			wantStatus: http.StatusGone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockDownloadCVProcess{processFunc: tt.processFunc})
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			h.Handle(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Handle() status = %v, wantStatus %v", w.Code, tt.wantStatus)
			}
		})
	}
}
