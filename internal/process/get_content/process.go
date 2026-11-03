package get_content

import (
	"context"
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
	var reference map[string]any

	for lang, filePath := range contentFiles {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not read content file for lang %s: %w", lang, err)
		}

		doc, err := decodeDocument(file)
		if err != nil {
			return nil, fmt.Errorf("content file for lang %s: %w", lang, err)
		}

		meta := doc["meta"].(map[string]any)

		if meta["locale"] != lang {
			return nil, fmt.Errorf("content file for lang %s has mismatched meta.locale", lang)
		}

		if reference == nil {
			reference = doc
		} else if err := compareStructure(reference, doc, "content"); err != nil {
			return nil, fmt.Errorf("content language parity: %w", err)
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
