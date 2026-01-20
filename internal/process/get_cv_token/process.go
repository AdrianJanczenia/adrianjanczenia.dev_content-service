package get_cv_token

import (
	"context"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type VerifyCaptchaTask interface {
	Execute(ctx context.Context, captchaID string) error
}

type ValidatePasswordTask interface {
	Execute(ctx context.Context, password, captchaID string) error
}

type DeleteCaptchaTask interface {
	Execute(ctx context.Context, captchaID string) error
}

type CreateTokenTask interface {
	Execute(ctx context.Context) (string, error)
}

type Process struct {
	verifyCaptchaTask    VerifyCaptchaTask
	validatePasswordTask ValidatePasswordTask
	deleteCaptchaTask    DeleteCaptchaTask
	createTokenTask      CreateTokenTask
	cvFilePaths          map[string]string
}

func NewProcess(verifyCaptchaTask VerifyCaptchaTask, validatePasswordTask ValidatePasswordTask, deleteCaptchaTask DeleteCaptchaTask, createTokenTask CreateTokenTask, cvFilePaths map[string]string) *Process {
	return &Process{
		verifyCaptchaTask:    verifyCaptchaTask,
		validatePasswordTask: validatePasswordTask,
		deleteCaptchaTask:    deleteCaptchaTask,
		createTokenTask:      createTokenTask,
		cvFilePaths:          cvFilePaths,
	}
}

func (p *Process) Process(ctx context.Context, password, lang, captchaID string) (string, error) {
	if _, ok := p.cvFilePaths[lang]; !ok {
		return "", errors.ErrUnsupportedLanguage
	}

	if err := p.verifyCaptchaTask.Execute(ctx, captchaID); err != nil {
		return "", err
	}

	if err := p.validatePasswordTask.Execute(ctx, password, captchaID); err != nil {
		return "", err
	}

	if err := p.deleteCaptchaTask.Execute(ctx, captchaID); err != nil {
		return "", err
	}

	token, err := p.createTokenTask.Execute(ctx)
	if err != nil {
		return "", err
	}

	return token, nil
}
