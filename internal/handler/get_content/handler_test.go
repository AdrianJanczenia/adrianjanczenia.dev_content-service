package get_content

import (
	"context"
	"errors"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/api/proto/v1"
	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockGetContentProcess struct {
	processFunc func(ctx context.Context, lang string) ([]byte, error)
}

func (m *mockGetContentProcess) Process(ctx context.Context, lang string) ([]byte, error) {
	return m.processFunc(ctx, lang)
}

func TestHandler_GetContent(t *testing.T) {
	tests := []struct {
		name        string
		req         *contentv1.GetContentRequest
		processFunc func(context.Context, string) ([]byte, error)
		wantCode    codes.Code
		wantRes     []byte
	}{
		{
			name: "successful response",
			req:  &contentv1.GetContentRequest{Lang: "pl"},
			processFunc: func(ctx context.Context, l string) ([]byte, error) {
				return []byte(`{"ok": true}`), nil
			},
			wantCode: codes.OK,
			wantRes:  []byte(`{"ok": true}`),
		},
		{
			name: "content not found",
			req:  &contentv1.GetContentRequest{Lang: "fr"},
			processFunc: func(ctx context.Context, l string) ([]byte, error) {
				return nil, appErrors.ErrContentNotFound
			},
			wantCode: codes.NotFound,
			wantRes:  nil,
		},
		{
			name: "internal error",
			req:  &contentv1.GetContentRequest{Lang: "en"},
			processFunc: func(ctx context.Context, l string) ([]byte, error) {
				return nil, errors.New("fs error")
			},
			wantCode: codes.Internal,
			wantRes:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockGetContentProcess{processFunc: tt.processFunc})
			res, err := h.Handle(context.Background(), tt.req)

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Errorf("Handle() unexpected error: %v", err)
				}
				if string(res.JsonContent) != string(tt.wantRes) {
					t.Errorf("Handle() got = %v, want %v", string(res.JsonContent), string(tt.wantRes))
				}
			} else {
				st, ok := status.FromError(err)
				if !ok || st.Code() != tt.wantCode {
					t.Errorf("Handle() expected code %v, got %v", tt.wantCode, st.Code())
				}
			}
		})
	}
}
