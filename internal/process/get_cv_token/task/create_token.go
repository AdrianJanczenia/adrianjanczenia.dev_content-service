package task

import (
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
	"github.com/google/uuid"
)

type TokenService interface {
	SetToken(token string, value interface{}, ttl time.Duration) error
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

func (t *CreateTokenTask) Execute() (string, error) {
	token := uuid.New().String()

	err := t.tokenService.SetToken(token, "valid", t.tokenTTL)
	if err != nil {
		return "", errors.ErrInternalServerError
	}

	return token, nil
}
