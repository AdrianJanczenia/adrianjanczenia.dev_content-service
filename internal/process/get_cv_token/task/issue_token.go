package task

import (
	"context"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/cv_state"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type TokenIssuer interface {
	IssueToken(ctx context.Context, captchaID, lang, candidate string, passwordCorrect bool, ttl time.Duration) (cv_state.IssueStatus, string, error)
}

type TokenGenerator interface {
	Execute() (string, error)
}

type IssueTokenTask struct {
	issuer    TokenIssuer
	generator TokenGenerator
	tokenTTL  time.Duration
}

func NewIssueTokenTask(issuer TokenIssuer, generator TokenGenerator, tokenTTL time.Duration) *IssueTokenTask {
	return &IssueTokenTask{issuer: issuer, generator: generator, tokenTTL: tokenTTL}
}

func (t *IssueTokenTask) Execute(ctx context.Context, captchaID, lang string, passwordCorrect bool) (string, error) {
	for range 3 {
		candidate, err := t.generator.Execute()
		if err != nil {
			return "", errors.ErrInternalServerError
		}

		status, token, err := t.issuer.IssueToken(ctx, captchaID, lang, candidate, passwordCorrect, t.tokenTTL)
		if err != nil {
			return "", errors.ErrInternalServerError
		}

		switch status {
		case cv_state.Issued, cv_state.Reused:
			return token, nil
		case cv_state.Collision:
			continue
		case cv_state.CaptchaNotFound:
			return "", errors.ErrCaptchaNotFound
		case cv_state.CaptchaNotSolved:
			return "", errors.ErrCaptchaNotSolved
		case cv_state.InvalidPassword:
			return "", errors.ErrInvalidPassword
		case cv_state.NoTriesLeft:
			return "", errors.ErrNoTriesLeft
		case cv_state.UnsupportedLanguage:
			return "", errors.ErrUnsupportedLanguage
		default:
			return "", errors.ErrInternalServerError
		}
	}
	return "", errors.ErrInternalServerError
}
