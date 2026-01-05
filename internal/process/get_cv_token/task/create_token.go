package task

import (
	"context"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
	"github.com/google/uuid"
)

type TokenService interface {
	SetToken(ctx context.Context, token string, value interface{}, ttl time.Duration) error
}

type CreateTokenTask struct {
	tokenService TokenService
	tokenTTL     time.Duration
}

func NewCreateTokenTask(ts TokenService, ttl time.Duration) *CreateTokenTask {
	return &CreateTokenTask{
		tokenService: ts,
		tokenTTL:     ttl,
	}
}

func (t *CreateTokenTask) Execute(ctx context.Context) (string, error) {
	token := uuid.New().String()

	err := t.tokenService.SetToken(ctx, token, "valid", t.tokenTTL)
	if err != nil {
		return "", errors.ErrInternalServerError
	}

	return token, nil
}
