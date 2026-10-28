//go:build integration

package redis

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/cv_state"
	"github.com/google/uuid"
)

func TestTokenScriptsAgainstRedis(t *testing.T) {
	redisURL := os.Getenv("REDIS_TEST_URL")
	if redisURL == "" {
		t.Skip("REDIS_TEST_URL is not set")
	}
	client, err := NewClient(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	id := uuid.NewString()
	firstToken := uuid.NewString()
	secondToken := uuid.NewString()
	captchaKey := "captcha:" + id
	resultKey := "cv:result:" + id
	firstTokenKey := "cv:token:" + firstToken
	secondTokenKey := "cv:token:" + secondToken
	defer client.rdb.Del(ctx, captchaKey, resultKey, firstTokenKey, secondTokenKey)
	if err := client.rdb.Set(ctx, captchaKey, `{"value":"ABC","triesLeft":3,"solved":true}`, 5*time.Second).Err(); err != nil {
		t.Fatal(err)
	}
	ttlBefore := client.rdb.PTTL(ctx, captchaKey).Val()

	status, _, err := client.IssueToken(ctx, id, "pl", firstToken, false, 5*time.Second)
	if err != nil || status != "invalid_password" {
		t.Fatalf("wrong password: status=%q err=%v", status, err)
	}
	stored, err := client.rdb.Get(ctx, captchaKey).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	var state struct {
		TriesLeft int `json:"triesLeft"`
	}
	if err := json.Unmarshal(stored, &state); err != nil || state.TriesLeft != 2 {
		t.Fatalf("wrong password did not decrement attempts: tries=%d err=%v", state.TriesLeft, err)
	}
	if ttlAfter := client.rdb.PTTL(ctx, captchaKey).Val(); ttlAfter <= 0 || ttlAfter > ttlBefore {
		t.Fatalf("CAPTCHA TTL changed unexpectedly: before=%v after=%v", ttlBefore, ttlAfter)
	}

	status, token, err := client.IssueToken(ctx, id, "pl", firstToken, true, 5*time.Second)
	if err != nil || status != "issued" || token != firstToken {
		t.Fatalf("issue: status=%q token=%q err=%v", status, token, err)
	}
	status, token, err = client.IssueToken(ctx, id, "pl", secondToken, true, 5*time.Second)
	if err != nil || status != "reused" || token != firstToken {
		t.Fatalf("retry: status=%q token=%q err=%v", status, token, err)
	}
	if client.rdb.Exists(ctx, secondTokenKey).Val() != 0 {
		t.Fatal("retry created a second token")
	}

	redeemStatus, err := client.RedeemToken(ctx, firstToken, "en")
	if err != nil || redeemStatus != cv_state.WrongLanguage {
		t.Fatalf("wrong language: status=%q err=%v", redeemStatus, err)
	}
	redeemStatus, err = client.RedeemToken(ctx, firstToken, "pl")
	if err != nil || redeemStatus != cv_state.Redeemed {
		t.Fatalf("redeem: status=%q err=%v", redeemStatus, err)
	}
	redeemStatus, err = client.RedeemToken(ctx, firstToken, "pl")
	if err != nil || redeemStatus != cv_state.TokenNotFound {
		t.Fatalf("second redemption: status=%q err=%v", redeemStatus, err)
	}
	if client.rdb.Exists(ctx, resultKey).Val() != 0 {
		t.Fatal("redemption left the idempotency key")
	}
}

func TestConcurrentTokenIssuanceAgainstRedis(t *testing.T) {
	redisURL := os.Getenv("REDIS_TEST_URL")
	if redisURL == "" {
		t.Skip("REDIS_TEST_URL is not set")
	}
	client, err := NewClient(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()
	id := uuid.NewString()
	const workers = 16
	candidates := make([]string, workers)
	keys := []string{"captcha:" + id, "cv:result:" + id}
	for i := range candidates {
		candidates[i] = uuid.NewString()
		keys = append(keys, "cv:token:"+candidates[i])
	}
	defer client.rdb.Del(ctx, keys...)
	if err := client.rdb.Set(ctx, "captcha:"+id, `{"value":"ABC","triesLeft":3,"solved":true}`, 5*time.Second).Err(); err != nil {
		t.Fatal(err)
	}
	type result struct {
		status cv_state.IssueStatus
		token  string
		err    error
	}
	results := make(chan result, workers)
	var wg sync.WaitGroup
	for _, candidate := range candidates {
		wg.Add(1)
		go func(candidate string) {
			defer wg.Done()
			status, token, err := client.IssueToken(ctx, id, "pl", candidate, true, 5*time.Second)
			results <- result{status: status, token: token, err: err}
		}(candidate)
	}
	wg.Wait()
	close(results)
	issued := 0
	var firstToken string
	for result := range results {
		if result.err != nil || (result.status != "issued" && result.status != "reused") || result.token == "" {
			t.Fatalf("concurrent issue = %+v", result)
		}
		if result.status == "issued" {
			issued++
		}
		if firstToken == "" {
			firstToken = result.token
		} else if result.token != firstToken {
			t.Fatalf("concurrent requests returned different tokens: %q and %q", firstToken, result.token)
		}
	}
	if issued != 1 {
		t.Fatalf("token was issued %d times, expected once", issued)
	}
}

func TestConcurrentTokenRedemptionAgainstRedis(t *testing.T) {
	redisURL := os.Getenv("REDIS_TEST_URL")
	if redisURL == "" {
		t.Skip("REDIS_TEST_URL is not set")
	}
	client, err := NewClient(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := context.Background()
	id := uuid.NewString()
	token := uuid.NewString()
	captchaKey := "captcha:" + id
	resultKey := "cv:result:" + id
	tokenKey := "cv:token:" + token
	defer client.rdb.Del(ctx, captchaKey, resultKey, tokenKey)
	if err := client.rdb.Set(ctx, captchaKey, `{"value":"ABC","triesLeft":3,"solved":true}`, 5*time.Second).Err(); err != nil {
		t.Fatal(err)
	}
	if status, _, err := client.IssueToken(ctx, id, "pl", token, true, 5*time.Second); err != nil || status != cv_state.Issued {
		t.Fatalf("issue token: status=%q err=%v", status, err)
	}

	const workers = 16
	results := make(chan cv_state.RedeemStatus, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, err := client.RedeemToken(ctx, token, "pl")
			if err != nil {
				results <- "error"
				return
			}
			results <- status
		}()
	}
	wg.Wait()
	close(results)
	counts := map[cv_state.RedeemStatus]int{}
	for status := range results {
		counts[status]++
	}
	if counts[cv_state.Redeemed] != 1 || counts[cv_state.TokenNotFound] != workers-1 || client.rdb.Exists(ctx, resultKey).Val() != 0 {
		t.Fatalf("unexpected concurrent redemption results: %+v", counts)
	}
}
