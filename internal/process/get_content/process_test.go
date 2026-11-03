package get_content

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func bundled(t *testing.T, lang string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "content", lang+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func writeContent(t *testing.T, name string, raw []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name+".json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBundledContent(t *testing.T) {
	p, err := NewProcess(map[string]string{"pl": filepath.Join("..", "..", "..", "content", "pl.json"), "en": filepath.Join("..", "..", "..", "content", "en.json")}, "pl")
	if err != nil {
		t.Fatal(err)
	}
	for _, lang := range []string{"pl", "en"} {
		raw, err := p.Process(context.Background(), lang)
		if err != nil {
			t.Fatal(err)
		}
		var d map[string]any
		if err := json.Unmarshal(raw, &d); err != nil {
			t.Fatal(err)
		}
		if d["schema_version"] != float64(2) {
			t.Fatalf("%s is not v2", lang)
		}
	}
	raw, err := p.Process(context.Background(), "de")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"locale": "pl"`) {
		t.Fatal("unknown language did not fall back to PL")
	}
}

func TestNewProcessRejectsBrokenDocuments(t *testing.T) {
	valid := bundled(t, "pl")
	var d map[string]any
	if err := json.Unmarshal(valid, &d); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		change func(map[string]any)
	}{
		{"unknown field", func(m map[string]any) { m["surprise"] = "text" }},
		{"wrong version", func(m map[string]any) { m["schema_version"] = float64(3) }},
		{"missing ordered case", func(m map[string]any) {
			m["pages"].(map[string]any)["experience"].(map[string]any)["cases"].(map[string]any)["order"] = []any{"migration", "checkout", "missing"}
		}},
		{"missing error copy", func(m map[string]any) {
			delete(m["ui"].(map[string]any)["errors"].(map[string]any), "error_cv_auth")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var copy map[string]any
			if err := json.Unmarshal(valid, &copy); err != nil {
				t.Fatal(err)
			}
			tt.change(copy)
			raw, err := json.Marshal(copy)
			if err != nil {
				t.Fatal(err)
			}
			_, err = NewProcess(map[string]string{"pl": writeContent(t, "pl", raw)}, "pl")
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	if _, err := NewProcess(map[string]string{"pl": writeContent(t, "pl", []byte("invalid"))}, "pl"); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	if _, err := NewProcess(map[string]string{"pl": writeContent(t, "pl", valid)}, "en"); err == nil {
		t.Fatal("missing default language accepted")
	}
}

func TestLanguageParity(t *testing.T) {
	pl := writeContent(t, "pl", bundled(t, "pl"))
	enRaw := bundled(t, "en")
	var en map[string]any
	if err := json.Unmarshal(enRaw, &en); err != nil {
		t.Fatal(err)
	}
	pages := en["pages"].(map[string]any)
	experience := pages["experience"].(map[string]any)
	cases := experience["cases"].(map[string]any)
	cases["order"] = []any{"checkout", "migration", "reservation"}
	changed, err := json.Marshal(en)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewProcess(map[string]string{"pl": pl, "en": writeContent(t, "en", changed)}, "pl")
	if err == nil || !strings.Contains(err.Error(), "different ID order") {
		t.Fatalf("different case order should be rejected, got %v", err)
	}
}

func TestLocalizedListsMayHaveDifferentLengths(t *testing.T) {
	pl := writeContent(t, "pl", bundled(t, "pl"))
	var en map[string]any
	if err := json.Unmarshal(bundled(t, "en"), &en); err != nil {
		t.Fatal(err)
	}
	pages := en["pages"].(map[string]any)
	experience := pages["experience"].(map[string]any)
	cases := experience["cases"].(map[string]any)
	items := cases["items"].(map[string]any)
	checkout := items["checkout"].(map[string]any)
	checkout["tech_context"] = append(checkout["tech_context"].([]any), "Another technology")
	changed, err := json.Marshal(en)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewProcess(map[string]string{"pl": pl, "en": writeContent(t, "en", changed)}, "pl"); err != nil {
		t.Fatalf("localized list length should not affect content parity: %v", err)
	}
}
