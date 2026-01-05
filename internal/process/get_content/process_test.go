package get_content

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewProcess(t *testing.T) {
	tmpDir := t.TempDir()
	plPath := filepath.Join(tmpDir, "pl.json")
	os.WriteFile(plPath, []byte(`{"hello": "cześć"}`), 0644)

	tests := []struct {
		name         string
		contentFiles map[string]string
		defaultLang  string
		wantErr      bool
	}{
		{
			name:         "successful initialization",
			contentFiles: map[string]string{"pl": plPath},
			defaultLang:  "pl",
			wantErr:      false,
		},
		{
			name:         "missing file error",
			contentFiles: map[string]string{"en": "non_existent.json"},
			defaultLang:  "en",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProcess(tt.contentFiles, tt.defaultLang)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProcess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcess_GetContent(t *testing.T) {
	tmpDir := t.TempDir()
	plPath := filepath.Join(tmpDir, "pl.json")
	enPath := filepath.Join(tmpDir, "en.json")
	os.WriteFile(plPath, []byte("pl content"), 0644)
	os.WriteFile(enPath, []byte("en content"), 0644)

	p, _ := NewProcess(map[string]string{"pl": plPath, "en": enPath}, "en")

	tests := []struct {
		name    string
		lang    string
		want    string
		wantErr error
	}{
		{
			name:    "get existing language",
			lang:    "pl",
			want:    "pl content",
			wantErr: nil,
		},
		{
			name:    "fallback to default language",
			lang:    "de",
			want:    "en content",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Process(context.Background(), tt.lang)
			if err != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if string(got) != tt.want {
				t.Errorf("Process() got = %v, want %v", string(got), tt.want)
			}
		})
	}
}
