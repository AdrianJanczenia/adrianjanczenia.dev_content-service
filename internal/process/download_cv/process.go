package download_cv

import "fmt"

type TokenValidator interface {
	ValidateAndDeleteToken(token string) (bool, error)
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

func (p *Process) Process(token, lang string) (string, error) {
	valid, err := p.tokenValidator.ValidateAndDeleteToken(token)
	if err != nil {
		return "", fmt.Errorf("could not validate token: %w", err)
	}
	if !valid {
		return "", fmt.Errorf("invalid or expired token")
	}

	filePath, ok := p.cvFilePaths[lang]
	if !ok {
		return "", fmt.Errorf("no cv available for language %s", lang)
	}

	return filePath, nil
}
