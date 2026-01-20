package get_cv_token

import (
	"context"
	"errors"
	"testing"
)

type mockVerifyCaptchaTask struct {
	executeFunc func(ctx context.Context, id string) error
}

func (m *mockVerifyCaptchaTask) Execute(ctx context.Context, id string) error {
	return m.executeFunc(ctx, id)
}

type mockValidatePasswordTask struct {
	executeFunc func(ctx context.Context, password, captchaID string) error
}

func (m *mockValidatePasswordTask) Execute(ctx context.Context, password, captchaID string) error {
	return m.executeFunc(ctx, password, captchaID)
}

type mockDeleteCaptchaTask struct {
	executeFunc func(ctx context.Context, id string) error
}

func (m *mockDeleteCaptchaTask) Execute(ctx context.Context, id string) error {
	return m.executeFunc(ctx, id)
}

type mockCreateTokenTask struct {
	executeFunc func(ctx context.Context) (string, error)
}

func (m *mockCreateTokenTask) Execute(ctx context.Context) (string, error) {
	return m.executeFunc(ctx)
}

func TestProcess_Process(t *testing.T) {
	paths := map[string]string{"pl": "path"}
	tests := []struct {
		name                 string
		lang                 string
		verifyCaptchaFunc    func(context.Context, string) error
		validatePasswordFunc func(context.Context, string, string) error
		deleteCaptchaFunc    func(context.Context, string) error
		createTokenFunc      func(context.Context) (string, error)
		wantErr              bool
	}{
		{
			name:                 "success",
			lang:                 "pl",
			verifyCaptchaFunc:    func(ctx context.Context, id string) error { return nil },
			validatePasswordFunc: func(ctx context.Context, p, id string) error { return nil },
			deleteCaptchaFunc:    func(ctx context.Context, id string) error { return nil },
			createTokenFunc:      func(ctx context.Context) (string, error) { return "token", nil },
			wantErr:              false,
		},
		{
			name:    "unsupported lang",
			lang:    "en",
			wantErr: true,
		},
		{
			name:              "captcha fail",
			lang:              "pl",
			verifyCaptchaFunc: func(ctx context.Context, id string) error { return errors.New("fail") },
			wantErr:           true,
		},
		{
			name:                 "password fail",
			lang:                 "pl",
			verifyCaptchaFunc:    func(ctx context.Context, id string) error { return nil },
			validatePasswordFunc: func(ctx context.Context, p, id string) error { return errors.New("fail") },
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProcess(
				&mockVerifyCaptchaTask{executeFunc: tt.verifyCaptchaFunc},
				&mockValidatePasswordTask{executeFunc: tt.validatePasswordFunc},
				&mockDeleteCaptchaTask{executeFunc: tt.deleteCaptchaFunc},
				&mockCreateTokenTask{executeFunc: tt.createTokenFunc},
				paths,
			)
			_, err := p.Process(context.Background(), "pass", tt.lang, "id")
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
