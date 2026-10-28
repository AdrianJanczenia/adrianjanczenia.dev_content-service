package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/cv_state"
)

type issueScriptInput struct {
	captchaID       string
	language        string
	candidate       string
	passwordCorrect bool
	tokenTTL        time.Duration
	priorResult     string
	priorToken      string
}

func (in issueScriptInput) keys() []string {
	priorToken := in.candidate
	if in.priorToken != "" {
		priorToken = in.priorToken
	}
	return []string{
		"captcha:" + in.captchaID,
		"cv:result:" + in.captchaID,
		"cv:token:" + in.candidate,
		"cv:token:" + priorToken,
	}
}

func (in issueScriptInput) args() []any {
	passwordFlag := "0"
	if in.passwordCorrect {
		passwordFlag = "1"
	}
	return []any{
		passwordFlag,
		in.language,
		in.candidate,
		in.tokenTTL.Milliseconds(),
		in.captchaID,
		in.priorResult,
		in.priorToken,
	}
}

func (c *Client) IssueToken(ctx context.Context, captchaID, lang, candidate string, passwordCorrect bool, ttl time.Duration) (cv_state.IssueStatus, string, error) {
	if c.rdb == nil {
		return "", "", fmt.Errorf("redis script client is not initialized")
	}

	priorRaw, priorToken := "", ""
	for range 4 {
		input := issueScriptInput{
			captchaID:       captchaID,
			language:        lang,
			candidate:       candidate,
			passwordCorrect: passwordCorrect,
			tokenTTL:        ttl,
			priorResult:     priorRaw,
			priorToken:      priorToken,
		}
		result, err := issueTokenScript.Run(ctx, c.rdb, input.keys(), input.args()...).Slice()
		if err != nil {
			return "", "", err
		}

		status, token, err := parseIssueResult(result)
		if err != nil {
			return "", "", err
		}

		if status != "retry" {
			return cv_state.IssueStatus(status), token, nil
		}

		priorRaw, priorToken, err = parsePriorResult(token)
		if err != nil {
			return "", "", err
		}
	}

	return "", "", fmt.Errorf("token state changed repeatedly")
}

func parseIssueResult(result []any) (string, string, error) {
	if len(result) != 2 {
		return "", "", fmt.Errorf("unexpected token script result")
	}

	status, okStatus := result[0].(string)
	token, okToken := result[1].(string)
	if !okStatus || !okToken {
		return "", "", fmt.Errorf("invalid token script result")
	}

	return status, token, nil
}

func parsePriorResult(raw string) (string, string, error) {
	if raw == "" {
		return "", "", nil
	}

	separator := strings.IndexByte(raw, ':')
	if separator < 1 || separator == len(raw)-1 {
		return "", "", fmt.Errorf("invalid idempotency key value")
	}

	return raw, raw[separator+1:], nil
}
