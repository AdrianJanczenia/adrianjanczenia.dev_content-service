package task

import (
	"context"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/cv_state"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type TokenRedeemer interface {
	RedeemToken(ctx context.Context, token, lang string) (cv_state.RedeemStatus, error)
}

type RedeemTokenTask struct {
	redeemer TokenRedeemer
}

func NewRedeemTokenTask(redeemer TokenRedeemer) *RedeemTokenTask {
	return &RedeemTokenTask{redeemer: redeemer}
}

func (t *RedeemTokenTask) Execute(ctx context.Context, token, lang string) error {
	status, err := t.redeemer.RedeemToken(ctx, token, lang)
	if err != nil {
		return errors.ErrInternalServerError
	}

	switch status {
	case cv_state.Redeemed:
		return nil
	case cv_state.TokenNotFound:
		return errors.ErrCVExpired
	case cv_state.WrongLanguage:
		return errors.ErrUnsupportedLanguage
	default:
		return errors.ErrInternalServerError
	}
}
