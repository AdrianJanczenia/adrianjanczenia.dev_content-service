package get_content

import (
	"fmt"
	"os"
)

type Process struct {
	content     map[string][]byte
	defaultLang string
}

func NewProcess(contentFiles map[string]string, defaultLang string) (*Process, error) {
	content := make(map[string][]byte)

	for lang, filePath := range contentFiles {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not read content file for lang %s: %w", lang, err)
		}

		content[lang] = file
	}

	return &Process{
		content:     content,
		defaultLang: defaultLang,
	}, nil
}

func (p *Process) Process(lang string) ([]byte, error) {
	if content, ok := p.content[lang]; ok {
		return content, nil
	}
	if content, ok := p.content[p.defaultLang]; ok {
		return content, nil
	}

	return nil, fmt.Errorf("no content available for language %s or default language", lang)
}
