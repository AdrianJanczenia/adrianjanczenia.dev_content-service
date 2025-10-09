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
}

func NewProcess(vpt ValidatePasswordTask, ctt CreateTokenTask) *Process {
	return &Process{
		validatePasswordTask: vpt,
		createTokenTask:      ctt,
	}
}

func (p *Process) Process(password string) (string, error) {
	if err := p.validatePasswordTask.Execute(password); err != nil {
		return "", err
	}

	token, err := p.createTokenTask.Execute()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/download/cv?token=%s", token), nil
}
