package download_cv

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type mockDownloadCVProcess struct {
	processFunc func(ctx context.Context, token, lang string) (io.ReadCloser, int64, error)
}

func (m *mockDownloadCVProcess) Process(ctx context.Context, token, lang string) (io.ReadCloser, int64, error) {
	return m.processFunc(ctx, token, lang)
}

func TestHandler_DownloadCV(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		processFunc func(context.Context, string, string) (io.ReadCloser, int64, error)
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
			processFunc: func(ctx context.Context, t, l string) (io.ReadCloser, int64, error) {
				return nil, 0, errors.ErrCVExpired
			},
			wantStatus: http.StatusGone,
		},
		{
			name:   "successful PDF stream",
			method: http.MethodGet,
			url:    "/download/cv?token=abc&lang=pl",
			processFunc: func(context.Context, string, string) (io.ReadCloser, int64, error) {
				return io.NopCloser(strings.NewReader("%PDF")), 4, nil
			},
			wantStatus: http.StatusOK,
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
			if tt.wantStatus == http.StatusMethodNotAllowed && w.Header().Get("Allow") != http.MethodGet {
				t.Errorf("Allow = %q, want GET", w.Header().Get("Allow"))
			}
			if tt.name == "successful PDF stream" && (w.Body.String() != "%PDF" || w.Header().Get("Content-Length") != "4" || w.Header().Get("Cache-Control") != "no-store") {
				t.Fatalf("unexpected PDF response: body=%q, headers=%v", w.Body.String(), w.Header())
			}
		})
	}
}
