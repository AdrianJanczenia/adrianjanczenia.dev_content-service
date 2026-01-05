package task

import "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"

type ValidatePasswordTask struct {
	correctPassword string
}

func NewValidatePasswordTask(correctPassword string) *ValidatePasswordTask {
	return &ValidatePasswordTask{correctPassword: correctPassword}
}

func (t *ValidatePasswordTask) Execute(password string) error {
	if password != t.correctPassword {
		return errors.ErrInvalidPassword
	}

	return nil
}
