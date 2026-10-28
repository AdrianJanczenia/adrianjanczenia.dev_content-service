package task

import (
	"io"
	"os"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type OpenFileTask struct {
	cvFilePaths map[string]string
}

func NewOpenFileTask(cvFilePaths map[string]string) *OpenFileTask {
	return &OpenFileTask{cvFilePaths: cvFilePaths}
}

func (t *OpenFileTask) Execute(lang string) (io.ReadCloser, int64, error) {
	path, ok := t.cvFilePaths[lang]
	if !ok {
		return nil, 0, errors.ErrUnsupportedLanguage
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, 0, errors.ErrInternalServerError
	}

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, 0, errors.ErrInternalServerError
	}

	return file, info.Size(), nil
}
