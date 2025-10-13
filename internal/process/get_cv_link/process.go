package get_cv_link

import (
	"fmt"
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
		return "", fmt.Errorf("unsupported language for cv: %s", lang)
	}

	if err := p.validatePasswordTask.Execute(password); err != nil {
		return "", err
	}

	token, err := p.createTokenTask.Execute()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/download/cv?token=%s&lang=%s", token, lang), nil
}
