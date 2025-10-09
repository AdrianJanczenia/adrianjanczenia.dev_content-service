package get_content

import (
	"encoding/json"
	"fmt"
	"os"
)

type Process struct {
	content     map[string]map[string]interface{}
	defaultLang string
}

func NewProcess(contentFiles map[string]string, defaultLang string) (*Process, error) {
	content := make(map[string]map[string]interface{})

	for lang, filePath := range contentFiles {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not read content file for lang %s: %w", lang, err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(file, &data); err != nil {
			return nil, fmt.Errorf("could not unmarshal content for lang %s: %w", lang, err)
		}
		content[lang] = data
	}

	return &Process{
		content:     content,
		defaultLang: defaultLang,
	}, nil
}

func (p *Process) Process(lang string) (map[string]interface{}, error) {
	if content, ok := p.content[lang]; ok {
		return content, nil
	}
	if content, ok := p.content[p.defaultLang]; ok {
		return content, nil
	}

	return nil, fmt.Errorf("no content available for language %s or default language", lang)
}
