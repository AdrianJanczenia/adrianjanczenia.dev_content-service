package task

import (
	"context"
	"fmt"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type DeleteCaptchaRedisClient interface {
	DelToken(ctx context.Context, key string) error
}

type DeleteCaptchaTask struct {
	client DeleteCaptchaRedisClient
}

func NewDeleteCaptchaTask(c DeleteCaptchaRedisClient) *DeleteCaptchaTask {
	return &DeleteCaptchaTask{client: c}
}

func (t *DeleteCaptchaTask) Execute(ctx context.Context, captchaID string) error {
	key := fmt.Sprintf("captcha:%s", captchaID)

	err := t.client.DelToken(ctx, key)
	if err != nil {
		return errors.ErrInternalServerError
	}

	return nil
}
