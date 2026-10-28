package get_cv_token

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type fakeValidateRequest struct {
	err   error
	calls int
}

func (f *fakeValidateRequest) Execute(lang, captchaID string) error {
	f.calls++
	if lang != "pl" || captchaID != "captcha-id" {
		return appErrors.ErrInvalidInput
	}
	return f.err
}

type fakeComparePassword struct {
	calls int
}

func (f *fakeComparePassword) Execute(password string) bool {
	f.calls++
	return password == "correct"
}

type fakeIssueToken struct {
	err             error
	calls           int
	passwordCorrect bool
}

func (f *fakeIssueToken) Execute(_ context.Context, captchaID, lang string, passwordCorrect bool) (string, error) {
	f.calls++
	f.passwordCorrect = passwordCorrect
	if captchaID != "captcha-id" || lang != "pl" {
		return "", appErrors.ErrInvalidInput
	}
	return "token", f.err
}

func TestProcessValidatesBeforeIssuing(t *testing.T) {
	validator := &fakeValidateRequest{err: appErrors.ErrUnsupportedLanguage}
	comparison := &fakeComparePassword{}
	issuer := &fakeIssueToken{}
	_, err := NewProcess(validator, comparison, issuer).Process(context.Background(), "correct", "pl", "captcha-id")
	if !errors.Is(err, appErrors.ErrUnsupportedLanguage) || validator.calls != 1 || comparison.calls != 0 || issuer.calls != 0 {
		t.Fatalf("validation should stop before password comparison and token issue: err=%v, calls=%d/%d/%d", err, validator.calls, comparison.calls, issuer.calls)
	}
}

func TestProcessPassesPasswordResultToIssuer(t *testing.T) {
	for _, tc := range []struct {
		password string
		correct  bool
	}{
		{password: "correct", correct: true},
		{password: "wrong", correct: false},
	} {
		validator := &fakeValidateRequest{}
		comparison := &fakeComparePassword{}
		issuer := &fakeIssueToken{}
		token, err := NewProcess(validator, comparison, issuer).Process(context.Background(), tc.password, "pl", "captcha-id")
		if err != nil || token != "token" || issuer.passwordCorrect != tc.correct || validator.calls != 1 || comparison.calls != 1 || issuer.calls != 1 {
			t.Fatalf("password %q: token=%q err=%v correct=%t calls=%d/%d/%d", tc.password, token, err, issuer.passwordCorrect, validator.calls, comparison.calls, issuer.calls)
		}
	}
}
