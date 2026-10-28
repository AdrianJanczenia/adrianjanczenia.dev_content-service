package task

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/cv_state"
	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

func TestOpenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cv.pdf")
	if err := os.WriteFile(path, []byte("%PDF"), 0600); err != nil {
		t.Fatal(err)
	}
	task := NewOpenFileTask(map[string]string{"pl": path})
	file, size, err := task.Execute("pl")
	if err != nil || size != 4 {
		t.Fatalf("open regular file: size=%d err=%v", size, err)
	}
	data, err := io.ReadAll(file)
	_ = file.Close()
	if err != nil || string(data) != "%PDF" {
		t.Fatalf("read file: data=%q err=%v", data, err)
	}
	if _, _, err := task.Execute("en"); err != appErrors.ErrUnsupportedLanguage {
		t.Fatalf("unsupported language: %v", err)
	}
	if _, _, err := NewOpenFileTask(map[string]string{"pl": path + ".missing"}).Execute("pl"); err != appErrors.ErrInternalServerError {
		t.Fatalf("missing file: %v", err)
	}
	if _, _, err := NewOpenFileTask(map[string]string{"pl": t.TempDir()}).Execute("pl"); err != appErrors.ErrInternalServerError {
		t.Fatalf("directory instead of file: %v", err)
	}
}

type fakeRedeemer struct {
	status cv_state.RedeemStatus
	err    error
}

func (f *fakeRedeemer) RedeemToken(_ context.Context, token, lang string) (cv_state.RedeemStatus, error) {
	if token != "token" || lang != "pl" {
		return "", errors.New("unexpected redemption arguments")
	}
	return f.status, f.err
}

func TestRedeemTokenMapsRedisOutcomes(t *testing.T) {
	for _, tc := range []struct {
		status cv_state.RedeemStatus
		err    error
		want   error
	}{
		{status: cv_state.Redeemed},
		{status: cv_state.TokenNotFound, want: appErrors.ErrCVExpired},
		{status: cv_state.WrongLanguage, want: appErrors.ErrUnsupportedLanguage},
		{status: cv_state.InvalidRedeemState, want: appErrors.ErrInternalServerError},
		{err: errors.New("redis unavailable"), want: appErrors.ErrInternalServerError},
	} {
		err := NewRedeemTokenTask(&fakeRedeemer{status: tc.status, err: tc.err}).Execute(context.Background(), "token", "pl")
		if err != tc.want {
			t.Fatalf("status %q, Redis err %v: got %v; want %v", tc.status, tc.err, err, tc.want)
		}
	}
}
