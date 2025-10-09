package download_cv

import "fmt"

type TokenValidator interface {
	ValidateAndDeleteToken(token string) (bool, error)
}

type Process struct {
	tokenValidator TokenValidator
	cvFilePath     string
}

func NewProcess(tv TokenValidator, cvPath string) *Process {
	return &Process{
		tokenValidator: tv,
		cvFilePath:     cvPath,
	}
}

func (p *Process) Process(token string) (string, error) {
	valid, err := p.tokenValidator.ValidateAndDeleteToken(token)
	if err != nil {
		return "", fmt.Errorf("could not validate token: %w", err)
	}
	if !valid {
		return "", fmt.Errorf("invalid or expired token")
	}
	return p.cvFilePath, nil
}
