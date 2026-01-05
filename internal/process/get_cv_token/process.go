package get_cv_token

import (
	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type ValidatePasswordTask interface {
	Execute(password string) error
}

type CreateTokenTask interface {
	Execute() (string, error)
}

type Process struct {
	validatePasswordTask ValidatePasswordTask
	createTokenTask      CreateTokenTask
	cvFilePaths          map[string]string
}

func NewProcess(vpt ValidatePasswordTask, ctt CreateTokenTask, cvPaths map[string]string) *Process {
	return &Process{
		validatePasswordTask: vpt,
		createTokenTask:      ctt,
		cvFilePaths:          cvPaths,
	}
}

func (p *Process) Process(password, lang string) (string, error) {
	if _, ok := p.cvFilePaths[lang]; !ok {
		return "", errors.ErrUnsupportedLanguage
	}

	if err := p.validatePasswordTask.Execute(password); err != nil {
		return "", err
	}

	token, err := p.createTokenTask.Execute()
	if err != nil {
		return "", err
	}

	return token, nil
}
