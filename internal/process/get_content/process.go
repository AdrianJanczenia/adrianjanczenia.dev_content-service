package get_content

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_content-service/internal/logic/errors"
)

type Process struct {
	content     map[string][]byte
	defaultLang string
}

func NewProcess(contentFiles map[string]string, defaultLang string) (*Process, error) {
	if defaultLang == "" || contentFiles[defaultLang] == "" {
		return nil, fmt.Errorf("default content language is not configured")
	}
	content := make(map[string][]byte)

	for lang, filePath := range contentFiles {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not read content file for lang %s: %w", lang, err)
		}
		if !json.Valid(file) {
			return nil, fmt.Errorf("content file for lang %s is not valid JSON", lang)
		}

		content[lang] = file
	}

	return &Process{
		content:     content,
		defaultLang: defaultLang,
	}, nil
}

func (p *Process) Process(ctx context.Context, lang string) ([]byte, error) {
	if content, ok := p.content[lang]; ok {
		return content, nil
	}
	if content, ok := p.content[p.defaultLang]; ok {
		return content, nil
	}

	return nil, errors.ErrContentNotFound
}
