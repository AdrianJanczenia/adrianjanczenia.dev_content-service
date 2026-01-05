package get_cv_token

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
	"github.com/rabbitmq/amqp091-go"
)

type mockGetCVTokenProcess struct {
	processFunc func(ctx context.Context, password, lang string) (string, error)
}

func (m *mockGetCVTokenProcess) Process(ctx context.Context, password, lang string) (string, error) {
	return m.processFunc(ctx, password, lang)
}

func TestHandler_GetCVToken(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		processFunc func(ctx context.Context, p, l string) (string, error)
		wantToken   string
		wantError   string
	}{
		{
			name: "valid request",
			body: []byte(`{"password": "pass", "lang": "pl"}`),
			processFunc: func(ctx context.Context, p, l string) (string, error) {
				return "token-123", nil
			},
			wantToken: "token-123",
			wantError: "",
		},
		{
			name: "process returns app error",
			body: []byte(`{"password": "wrong", "lang": "pl"}`),
			processFunc: func(ctx context.Context, p, l string) (string, error) {
				return "", appErrors.ErrInvalidPassword
			},
			wantToken: "",
			wantError: appErrors.ErrInvalidPassword.Slug,
		},
		{
			name: "process returns unknown error",
			body: []byte(`{"password": "pass", "lang": "pl"}`),
			processFunc: func(ctx context.Context, p, l string) (string, error) {
				return "", errors.New("db error")
			},
			wantToken: "",
			wantError: appErrors.ErrInternalServerError.Slug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockGetCVTokenProcess{processFunc: tt.processFunc})
			delivery := amqp091.Delivery{Body: tt.body}

			res, err := h.Handle(context.Background(), delivery)

			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}

			payload := res.(responsePayload)
			if payload.Token != tt.wantToken {
				t.Errorf("Handle() token = %v, wantToken %v", payload.Token, tt.wantToken)
			}
			if payload.Error != tt.wantError {
				t.Errorf("Handle() error = %v, wantError %v", payload.Error, tt.wantError)
			}
		})
	}
}
