package download_cv

import (
	"context"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type TokenValidator interface {
	ValidateAndDeleteToken(ctx context.Context, token string) (bool, error)
}

type Process struct {
	tokenValidator TokenValidator
	cvFilePaths    map[string]string
}

func NewProcess(tv TokenValidator, cvPaths map[string]string) *Process {
	return &Process{
		tokenValidator: tv,
		cvFilePaths:    cvPaths,
	}
}

func (p *Process) Process(ctx context.Context, token, lang string) (string, error) {
	valid, err := p.tokenValidator.ValidateAndDeleteToken(ctx, token)
	if err != nil {
		return "", errors.ErrInternalServerError
	}
	if !valid {
		return "", errors.ErrCVExpired
	}

	filePath, ok := p.cvFilePaths[lang]
	if !ok {
		return "", errors.ErrUnsupportedLanguage
	}

	return filePath, nil
}
