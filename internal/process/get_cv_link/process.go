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

type TokenValidator interface {
	ValidateToken(token string) (bool, error)
}

type Process struct {
	validatePasswordTask ValidatePasswordTask
	createTokenTask      CreateTokenTask
	tokenValidator       TokenValidator
	cvFilePath           string
}

func NewProcess(vpt ValidatePasswordTask, ctt CreateTokenTask, tv TokenValidator, cvPath string) *Process {
	return &Process{
		validatePasswordTask: vpt,
		createTokenTask:      ctt,
		tokenValidator:       tv,
		cvFilePath:           cvPath,
	}
}

func (p *Process) GenerateLink(password string) (string, error) {
	if err := p.validatePasswordTask.Execute(password); err != nil {
		return "", err
	}

	token, err := p.createTokenTask.Execute()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/download/cv?token=%s", token), nil
}

func (p *Process) ValidateTokenAndGetPath(token string) (string, error) {
	valid, err := p.tokenValidator.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("could not validate token: %w", err)
	}
	if !valid {
		return "", fmt.Errorf("invalid or expired token")
	}
	return p.cvFilePath, nil
}
