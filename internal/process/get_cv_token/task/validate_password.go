package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type ValidatePasswordRedisClient interface {
	GetToken(ctx context.Context, key string) (string, error)
	SetToken(ctx context.Context, token string, value interface{}, ttl time.Duration) error
	DelToken(ctx context.Context, key string) error
}

type captcha struct {
	Value     string `json:"value"`
	TriesLeft int    `json:"triesLeft"`
	Solved    bool   `json:"solved"`
}

type ValidatePasswordTask struct {
	correctPassword   string
	client            ValidatePasswordRedisClient
	captchaTtlMinutes int
}

func NewValidatePasswordTask(correctPassword string, client ValidatePasswordRedisClient, captchaTTL int) *ValidatePasswordTask {
	return &ValidatePasswordTask{
		correctPassword:   correctPassword,
		client:            client,
		captchaTtlMinutes: captchaTTL,
	}
}

func (t *ValidatePasswordTask) Execute(ctx context.Context, password, captchaID string) error {
	if password == t.correctPassword {
		return nil
	}

	key := fmt.Sprintf("captcha:%s", captchaID)

	data, err := t.client.GetToken(ctx, key)
	if err != nil {
		return errors.ErrCaptchaNotFound
	}

	var c captcha
	if err = json.Unmarshal([]byte(data), &c); err != nil {
		return errors.ErrInternalServerError
	}

	c.TriesLeft--

	if c.TriesLeft <= 0 {
		err = t.client.DelToken(ctx, key)
		if err != nil {
			return errors.ErrInternalServerError
		}

		return errors.ErrNoTriesLeft
	}

	newData, _ := json.Marshal(c)
	err = t.client.SetToken(ctx, key, string(newData), time.Duration(t.captchaTtlMinutes)*time.Minute)
	if err != nil {
		return errors.ErrInternalServerError
	}

	return errors.ErrInvalidPassword
}
