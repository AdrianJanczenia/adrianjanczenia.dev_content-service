package download_cv

import (
	"context"
	"io"
)

type OpenFileTask interface {
	Execute(lang string) (io.ReadCloser, int64, error)
}

type RedeemTokenTask interface {
	Execute(ctx context.Context, token, lang string) error
}

type Process struct {
	openFileTask    OpenFileTask
	redeemTokenTask RedeemTokenTask
}

func NewProcess(openFileTask OpenFileTask, redeemTokenTask RedeemTokenTask) *Process {
	return &Process{openFileTask: openFileTask, redeemTokenTask: redeemTokenTask}
}

func (p *Process) Process(ctx context.Context, token, lang string) (io.ReadCloser, int64, error) {
	file, size, err := p.openFileTask.Execute(lang)
	if err != nil {
		return nil, 0, err
	}

	if err := p.redeemTokenTask.Execute(ctx, token, lang); err != nil {
		_ = file.Close()
		return nil, 0, err
	}

	return file, size, nil
}
