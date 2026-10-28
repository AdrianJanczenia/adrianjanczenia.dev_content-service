package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

func TestClientPing(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{rdb: db}
	mock.ExpectPing().SetVal("PONG")
	if err := client.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIssueToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{rdb: db}
	keys := []string{"captcha:session", "cv:result:session", "cv:token:candidate", "cv:token:candidate"}
	mock.ExpectEvalSha(issueTokenScript.Hash(), keys, "1", "pl", "candidate", int64(60000), "session", "", "").SetVal([]interface{}{"issued", "candidate"})
	status, token, err := client.IssueToken(context.Background(), "session", "pl", "candidate", true, time.Minute)
	if err != nil || status != "issued" || token != "candidate" {
		t.Fatalf("IssueToken() = %q, %q, %v", status, token, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIssueTokenRetriesWithDeclaredPriorKey(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{rdb: db}
	initialKeys := []string{"captcha:session", "cv:result:session", "cv:token:candidate", "cv:token:candidate"}
	priorKeys := []string{"captcha:session", "cv:result:session", "cv:token:candidate", "cv:token:original"}
	mock.ExpectEvalSha(issueTokenScript.Hash(), initialKeys, "1", "pl", "candidate", int64(60000), "session", "", "").SetVal([]interface{}{"retry", "pl:original"})
	mock.ExpectEvalSha(issueTokenScript.Hash(), priorKeys, "1", "pl", "candidate", int64(60000), "session", "pl:original", "original").SetVal([]interface{}{"reused", "original"})
	status, token, err := client.IssueToken(context.Background(), "session", "pl", "candidate", true, time.Minute)
	if err != nil || status != "reused" || token != "original" {
		t.Fatalf("IssueToken retry = %q, %q, %v", status, token, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRedeemToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{rdb: db}
	mock.ExpectGet("cv:token:candidate").SetVal("pl:session")
	mock.ExpectEvalSha(redeemTokenScript.Hash(), []string{"cv:token:candidate", "cv:result:session"}, "pl", "pl:session").SetVal("redeemed")
	status, err := client.RedeemToken(context.Background(), "candidate", "pl")
	if err != nil || status != "redeemed" {
		t.Fatalf("RedeemToken() = %q, %v", status, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRedeemTokenMissing(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &Client{rdb: db}
	mock.ExpectGet("cv:token:missing").SetErr(redis.Nil)
	status, err := client.RedeemToken(context.Background(), "missing", "pl")
	if err != nil || status != "not_found" {
		t.Fatalf("RedeemToken missing = %q, %v", status, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
