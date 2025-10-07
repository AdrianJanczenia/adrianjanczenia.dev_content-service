package task

import "errors"

type ValidatePasswordTask struct {
	correctPassword string
}

func NewValidatePasswordTask(correctPassword string) *ValidatePasswordTask {
	return &ValidatePasswordTask{correctPassword: correctPassword}
}

func (t *ValidatePasswordTask) Execute(password string) error {
	if password != t.correctPassword {
		return errors.New("invalid password")
	}
	return nil
}
