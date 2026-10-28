package download_cv

import (
	"context"
	"errors"
	"io"
	"testing"

	appErrors "github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type trackedFile struct {
	closed bool
}

func (f *trackedFile) Read([]byte) (int, error) { return 0, io.EOF }
func (f *trackedFile) Close() error             { f.closed = true; return nil }

type fakeOpenFile struct {
	file  *trackedFile
	err   error
	calls int
}

func (f *fakeOpenFile) Execute(lang string) (io.ReadCloser, int64, error) {
	f.calls++
	if lang != "pl" {
		return nil, 0, appErrors.ErrUnsupportedLanguage
	}
	return f.file, 4, f.err
}

type fakeRedeemToken struct {
	err   error
	calls int
}

func (f *fakeRedeemToken) Execute(_ context.Context, token, lang string) error {
	f.calls++
	if token != "token" || lang != "pl" {
		return appErrors.ErrInvalidInput
	}
	return f.err
}

func TestProcessKeepsFileOpenAfterRedemption(t *testing.T) {
	file := &trackedFile{}
	opener := &fakeOpenFile{file: file}
	redeemer := &fakeRedeemToken{}
	got, size, err := NewProcess(opener, redeemer).Process(context.Background(), "token", "pl")
	if err != nil || got != file || size != 4 || file.closed || opener.calls != 1 || redeemer.calls != 1 {
		t.Fatalf("unexpected result: file=%v size=%d err=%v closed=%t calls=%d/%d", got, size, err, file.closed, opener.calls, redeemer.calls)
	}
	_ = got.Close()
}

func TestProcessClosesFileWhenRedemptionFails(t *testing.T) {
	file := &trackedFile{}
	redeemer := &fakeRedeemToken{err: appErrors.ErrCVExpired}
	got, size, err := NewProcess(&fakeOpenFile{file: file}, redeemer).Process(context.Background(), "token", "pl")
	if !errors.Is(err, appErrors.ErrCVExpired) || got != nil || size != 0 || !file.closed || redeemer.calls != 1 {
		t.Fatalf("unexpected result: file=%v size=%d err=%v closed=%t calls=%d", got, size, err, file.closed, redeemer.calls)
	}
}

func TestProcessDoesNotRedeemWhenFileCannotOpen(t *testing.T) {
	redeemer := &fakeRedeemToken{}
	_, _, err := NewProcess(&fakeOpenFile{err: appErrors.ErrInternalServerError}, redeemer).Process(context.Background(), "token", "pl")
	if !errors.Is(err, appErrors.ErrInternalServerError) || redeemer.calls != 0 {
		t.Fatalf("file error consumed token: err=%v calls=%d", err, redeemer.calls)
	}
}
