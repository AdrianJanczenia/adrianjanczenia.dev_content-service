package get_cv_token

import (
	"context"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
	"github.com/rabbitmq/amqp091-go"
)

type mockGetCVTokenProcess struct {
	processFunc func(ctx context.Context, password, lang, captchaID string) (string, error)
}

func (m *mockGetCVTokenProcess) Process(ctx context.Context, p, l, c string) (string, error) {
	return m.processFunc(ctx, p, l, c)
}

func TestHandler_Handle(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		processFunc func(context.Context, string, string, string) (string, error)
		wantToken   string
		wantError   string
	}{
		{
			name: "success",
			body: `{"password":"p","lang":"pl","captchaId":"c"}`,
			processFunc: func(ctx context.Context, p, l, c string) (string, error) {
				return "t123", nil
			},
			wantToken: "t123",
		},
		{
			name: "process error",
			body: `{"password":"p","lang":"pl","captchaId":"c"}`,
			processFunc: func(ctx context.Context, p, l, c string) (string, error) {
				return "", appErrors.ErrInvalidPassword
			},
			wantError: "error_cv_auth",
		},
		{
			name:      "unmarshal error",
			body:      `invalid`,
			wantError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockGetCVTokenProcess{processFunc: tt.processFunc}
			h := NewHandler(m)

			d := amqp091.Delivery{Body: []byte(tt.body)}
			res, err := h.Handle(context.Background(), d)

			if tt.body == "invalid" {
				if err != appErrors.ErrInvalidInput {
					t.Errorf("Handle() expected error_message")
				}
				return
			}

			payload := res.(responsePayload)
			if payload.Token != tt.wantToken {
				t.Errorf("Handle() got token = %v, want %v", payload.Token, tt.wantToken)
			}
			if payload.Error != tt.wantError {
				t.Errorf("Handle() got error = %v, want %v", payload.Error, tt.wantError)
			}
		})
	}
}
