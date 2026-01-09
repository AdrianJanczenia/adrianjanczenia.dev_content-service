package task

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
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
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	b := make([]byte, 32)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", errors.ErrInternalServerError
		}

		b[i] = charset[num.Int64()]
	}

	token := string(b)

	err := t.tokenService.SetToken(ctx, token, "valid", t.tokenTTL)
	if err != nil {
		return "", errors.ErrInternalServerError
	}

	return token, nil
}
