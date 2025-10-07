package get_content

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Process struct {
	content map[string]map[string]interface{}
}

func NewProcess(contentPath string) (*Process, error) {
	content := make(map[string]map[string]interface{})
	supportedLangs := []string{"pl", "en"}

	for _, lang := range supportedLangs {
		filePath := filepath.Join(contentPath, fmt.Sprintf("%s.json", lang))
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

	return &Process{content: content}, nil
}

func (p *Process) Execute(lang string) (map[string]interface{}, error) {
	if content, ok := p.content[lang]; ok {
		return content, nil
	}
	return p.content["pl"], nil
}
