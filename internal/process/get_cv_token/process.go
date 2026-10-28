package get_cv_token

import "context"

type ValidateRequestTask interface {
	Execute(lang, captchaID string) error
}

type ComparePasswordTask interface {
	Execute(password string) bool
}

type IssueTokenTask interface {
	Execute(ctx context.Context, captchaID, lang string, passwordCorrect bool) (string, error)
}

type Process struct {
	validateRequestTask ValidateRequestTask
	comparePasswordTask ComparePasswordTask
	issueTokenTask      IssueTokenTask
}

func NewProcess(validateRequestTask ValidateRequestTask, comparePasswordTask ComparePasswordTask, issueTokenTask IssueTokenTask) *Process {
	return &Process{
		validateRequestTask: validateRequestTask,
		comparePasswordTask: comparePasswordTask,
		issueTokenTask:      issueTokenTask,
	}
}

func (p *Process) Process(ctx context.Context, password, lang, captchaID string) (string, error) {
	if err := p.validateRequestTask.Execute(lang, captchaID); err != nil {
		return "", err
	}

	passwordCorrect := p.comparePasswordTask.Execute(password)

	return p.issueTokenTask.Execute(ctx, captchaID, lang, passwordCorrect)
}
