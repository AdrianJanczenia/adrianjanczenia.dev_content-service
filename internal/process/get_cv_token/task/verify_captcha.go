package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type VerifyCaptchaRedisClient interface {
	GetToken(ctx context.Context, key string) (string, error)
}

type VerifyCaptchaTask struct {
	client VerifyCaptchaRedisClient
}

func NewVerifyCaptchaTask(c VerifyCaptchaRedisClient) *VerifyCaptchaTask {
	return &VerifyCaptchaTask{client: c}
}

func (t *VerifyCaptchaTask) Execute(ctx context.Context, captchaID string) error {
	key := fmt.Sprintf("captcha:%s", captchaID)

	data, err := t.client.GetToken(ctx, key)
	if err != nil {
		return errors.ErrCaptchaNotFound
	}

	var c struct {
		Solved bool `json:"solved"`
	}

	if err := json.Unmarshal([]byte(data), &c); err != nil {
		return errors.ErrInternalServerError
	}

	if !c.Solved {
		return errors.ErrCaptchaNotSolved
	}

	return nil
}
