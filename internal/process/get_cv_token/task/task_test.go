package task

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/cv_state"
	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

func TestValidateRequest(t *testing.T) {
	task := NewValidateRequestTask(map[string]string{"pl": "cv.pdf"})
	for _, tc := range []struct {
		lang string
		id   string
		want error
	}{
		{lang: "pl", id: "captcha"},
		{lang: "en", id: "captcha", want: appErrors.ErrUnsupportedLanguage},
		{lang: "pl", want: appErrors.ErrCaptchaNotFound},
	} {
		if got := task.Execute(tc.lang, tc.id); got != tc.want {
			t.Fatalf("Execute(%q, %q) = %v; want %v", tc.lang, tc.id, got, tc.want)
		}
	}
}

func TestComparePassword(t *testing.T) {
	task := NewComparePasswordTask("correct")
	if !task.Execute("correct") || task.Execute("wrong") {
		t.Fatal("password comparison returned an incorrect result")
	}
}

func TestCreateToken(t *testing.T) {
	task := NewCreateTokenTask()
	first, err := task.Execute()
	if err != nil {
		t.Fatal(err)
	}
	second, err := task.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]{32}$`).MatchString(first) || first == second {
		t.Fatal("generated tokens have an invalid format or are identical")
	}
}

type fakeGenerator struct {
	calls int
	err   error
}

func (f *fakeGenerator) Execute() (string, error) {
	f.calls++
	return "candidate", f.err
}

type fakeIssuer struct {
	statuses []cv_state.IssueStatus
	result   string
	err      error
	calls    int
}

func (f *fakeIssuer) IssueToken(_ context.Context, id, lang, candidate string, correct bool, ttl time.Duration) (cv_state.IssueStatus, string, error) {
	f.calls++
	if id != "captcha" || lang != "pl" || candidate != "candidate" || !correct || ttl != time.Minute {
		return "", "", errors.New("unexpected issue arguments")
	}
	if f.err != nil {
		return "", "", f.err
	}
	return f.statuses[f.calls-1], f.result, nil
}

func TestIssueTokenMapsRedisOutcomes(t *testing.T) {
	for _, tc := range []struct {
		status cv_state.IssueStatus
		want   error
	}{
		{status: cv_state.Issued},
		{status: cv_state.Reused},
		{status: cv_state.CaptchaNotFound, want: appErrors.ErrCaptchaNotFound},
		{status: cv_state.CaptchaNotSolved, want: appErrors.ErrCaptchaNotSolved},
		{status: cv_state.InvalidPassword, want: appErrors.ErrInvalidPassword},
		{status: cv_state.NoTriesLeft, want: appErrors.ErrNoTriesLeft},
		{status: cv_state.UnsupportedLanguage, want: appErrors.ErrUnsupportedLanguage},
		{status: cv_state.InvalidIssueState, want: appErrors.ErrInternalServerError},
	} {
		issuer := &fakeIssuer{statuses: []cv_state.IssueStatus{tc.status}, result: "token"}
		generator := &fakeGenerator{}
		token, err := NewIssueTokenTask(issuer, generator, time.Minute).Execute(context.Background(), "captcha", "pl", true)
		if err != tc.want || issuer.calls != 1 || generator.calls != 1 || (err == nil && token != "token") || (err != nil && token != "") {
			t.Fatalf("status %q: token=%q err=%v calls=%d/%d", tc.status, token, err, issuer.calls, generator.calls)
		}
	}
}

func TestIssueTokenRetriesCollisionAtMostThreeTimes(t *testing.T) {
	issuer := &fakeIssuer{statuses: []cv_state.IssueStatus{cv_state.Collision, cv_state.Collision, cv_state.Issued}, result: "token"}
	generator := &fakeGenerator{}
	token, err := NewIssueTokenTask(issuer, generator, time.Minute).Execute(context.Background(), "captcha", "pl", true)
	if err != nil || token != "token" || issuer.calls != 3 || generator.calls != 3 {
		t.Fatalf("retry result: token=%q err=%v calls=%d/%d", token, err, issuer.calls, generator.calls)
	}
	issuer = &fakeIssuer{statuses: []cv_state.IssueStatus{cv_state.Collision, cv_state.Collision, cv_state.Collision}}
	_, err = NewIssueTokenTask(issuer, &fakeGenerator{}, time.Minute).Execute(context.Background(), "captcha", "pl", true)
	if err != appErrors.ErrInternalServerError || issuer.calls != 3 {
		t.Fatalf("collision exhaustion: err=%v calls=%d", err, issuer.calls)
	}
}

func TestIssueTokenStopsOnDependencyErrors(t *testing.T) {
	issuer := &fakeIssuer{statuses: []cv_state.IssueStatus{cv_state.Issued}, err: errors.New("redis unavailable")}
	_, err := NewIssueTokenTask(issuer, &fakeGenerator{}, time.Minute).Execute(context.Background(), "captcha", "pl", true)
	if err != appErrors.ErrInternalServerError || issuer.calls != 1 {
		t.Fatalf("issuer error: err=%v calls=%d", err, issuer.calls)
	}
	issuer = &fakeIssuer{statuses: []cv_state.IssueStatus{cv_state.Issued}}
	_, err = NewIssueTokenTask(issuer, &fakeGenerator{err: errors.New("random unavailable")}, time.Minute).Execute(context.Background(), "captcha", "pl", true)
	if err != appErrors.ErrInternalServerError || issuer.calls != 0 {
		t.Fatalf("generator error: err=%v calls=%d", err, issuer.calls)
	}
}
