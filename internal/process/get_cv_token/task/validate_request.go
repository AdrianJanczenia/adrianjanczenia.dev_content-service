package task

import "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"

type ValidateRequestTask struct {
	cvFilePaths map[string]string
}

func NewValidateRequestTask(cvFilePaths map[string]string) *ValidateRequestTask {
	return &ValidateRequestTask{cvFilePaths: cvFilePaths}
}

func (t *ValidateRequestTask) Execute(lang, captchaID string) error {
	if _, ok := t.cvFilePaths[lang]; !ok {
		return errors.ErrUnsupportedLanguage
	}

	if captchaID == "" {
		return errors.ErrCaptchaNotFound
	}

	return nil
}
