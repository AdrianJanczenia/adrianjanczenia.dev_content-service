package get_content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

var textBlock = object("eyebrow", "title", "body", "action")
var roleShape = map[string]any{"title": "", "period": "", "summary": "", "tasks": []any{""}}
var caseShape = map[string]any{"category": "", "title": "", "summary": "", "problem": "", "contribution": []any{""}, "result": "", "takeaway": "", "tech_context": []any{""}}
var techShape = map[string]any{"title": "", "items": []any{object("id", "label")}}
var orderedRoleShape = map[string]any{"order": []any{""}, "items": map[string]any{"*": roleShape}}
var orderedCaseShape = map[string]any{"order": []any{""}, "items": map[string]any{"*": caseShape}}
var orderedTechShape = map[string]any{"order": []any{""}, "items": map[string]any{"*": techShape}}
var documentShape = map[string]any{
	"schema_version": float64(0),
	"meta":           object("locale", "title", "description"),
	"shared":         map[string]any{"navigation": object("about", "experience", "how_it_works", "contact", "cv", "menu"), "footer": object("privacy"), "contact_links": object("email", "github", "linkedin")},
	"pages": map[string]any{
		"about":        map[string]any{"name": "", "role": "", "identity_detail": "", "tags": []any{""}, "hero": textBlock, "career_snapshot": textBlock, "site_teaser": textBlock},
		"experience":   map[string]any{"intro": textBlock, "roles": orderedRoleShape, "cases": orderedCaseShape, "technology_groups": orderedTechShape, "roles_label": "", "cases_label": "", "technology_label": "", "read_case": "", "back": "", "problem_label": "", "contribution_label": "", "result_label": ""},
		"how_it_works": map[string]any{"eyebrow": "", "title": "", "body": "", "ui_copy": map[string]any{"*": ""}, "scenario_copy": map[string]any{"*": map[string]any{"*": ""}}},
		"contact":      map[string]any{"intro": textBlock, "email_card": textBlock, "profile_card": textBlock},
		"privacy":      map[string]any{"title": "", "sections": []any{map[string]any{"header": "", "items": []any{object("label", "text")}}}},
	},
	"ui": map[string]any{"cv": object("title", "intro", "access_code_label", "captcha_label", "submit", "retry", "close", "contact_prompt"), "email_copied": "", "errors": map[string]any{"*": object("title", "message", "action")}},
}

func object(keys ...string) map[string]any {
	m := make(map[string]any, len(keys))
	for _, k := range keys {
		m[k] = ""
	}

	return m
}

func decodeDocument(raw []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()

	var doc map[string]any

	if err := dec.Decode(&doc); err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, fmt.Errorf("document must be an object")
	}

	if err := checkShape(doc, documentShape, "content"); err != nil {
		return nil, err
	}

	if doc["schema_version"] != float64(2) {
		return nil, fmt.Errorf("schema_version must be 2")
	}

	pages := doc["pages"].(map[string]any)
	exp := pages["experience"].(map[string]any)

	for _, section := range []string{"roles", "cases", "technology_groups"} {
		ordered := exp[section].(map[string]any)
		order := ordered["order"].([]any)
		items := ordered["items"].(map[string]any)
		seen := make(map[string]bool, len(order))

		for _, item := range order {
			id := item.(string)

			if id == "" || seen[id] {
				return nil, fmt.Errorf("%s has an empty or duplicate ID", section)
			}

			if _, ok := items[id]; !ok {
				return nil, fmt.Errorf("%s order references missing ID %s", section, id)
			}

			seen[id] = true
		}

		if len(seen) != len(items) {
			return nil, fmt.Errorf("%s items and order differ", section)
		}
	}
	errorsCopy := doc["ui"].(map[string]any)["errors"].(map[string]any)
	for _, slug := range []string{
		"error_cv_auth", "error_cv_expired", "error_cv_not_found", "error_cv_server",
		"error_internal", "error_message", "error_service_unavailable", "error_bad_gateway",
		"error_method_not_allowed", "error_invalid_input", "error_unsupported_language",
		"error_content_not_found", "error_captcha_server", "error_captcha_invalid",
		"error_captcha_not_solved", "error_captcha_expired", "error_captcha_not_found",
		"error_pow_failed", "error_pow_signature", "error_pow_work", "error_pow_double_spend",
		"error_pow_expired", "error_route_not_found",
	} {
		if _, ok := errorsCopy[slug]; !ok {
			return nil, fmt.Errorf("ui.errors.%s is missing", slug)
		}
	}

	return doc, nil
}

func checkShape(value, shape any, path string) error {
	switch expected := shape.(type) {
	case string:
		actual, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s must be a string", path)
		}
		if strings.TrimSpace(actual) == "" && !strings.HasSuffix(path, ".label") && path != "content.pages.experience.intro.action" && path != "content.pages.contact.intro.action" {
			return fmt.Errorf("%s must not be empty", path)
		}
	case float64:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%s must be a number", path)
		}
	case []any:
		list, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s must be an array", path)
		}
		if len(list) == 0 {
			return fmt.Errorf("%s must not be empty", path)
		}
		for i, item := range list {
			if err := checkShape(item, expected[0], fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	case map[string]any:
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s must be an object", path)
		}
		if len(object) == 0 {
			return fmt.Errorf("%s must not be empty", path)
		}
		if wildcard, ok := expected["*"]; ok {
			for key, item := range object {
				if strings.TrimSpace(key) == "" {
					return fmt.Errorf("%s has an empty ID", path)
				}
				if err := checkShape(item, wildcard, path+"."+key); err != nil {
					return err
				}
			}
			return nil
		}

		for key := range expected {
			if _, ok := object[key]; !ok {
				return fmt.Errorf("%s.%s is missing", path, key)
			}
		}

		for key, item := range object {
			child, ok := expected[key]
			if !ok {
				return fmt.Errorf("%s.%s is unknown", path, key)
			}
			if err := checkShape(item, child, path+"."+key); err != nil {
				return err
			}
		}
	}

	return nil
}

func compareStructure(a, b any, path string) error {
	switch left := a.(type) {
	case map[string]any:
		right, ok := b.(map[string]any)
		if !ok {
			return fmt.Errorf("%s has different type", path)
		}

		ak, bk := make([]string, 0, len(left)), make([]string, 0, len(right))

		for k := range left {
			ak = append(ak, k)
		}

		for k := range right {
			bk = append(bk, k)
		}

		sort.Strings(ak)
		sort.Strings(bk)
		if !reflect.DeepEqual(ak, bk) {
			return fmt.Errorf("%s has different keys/IDs", path)
		}

		for _, k := range ak {
			if err := compareStructure(left[k], right[k], path+"."+k); err != nil {
				return err
			}
		}
	case []any:
		right, ok := b.([]any)
		if !ok {
			return fmt.Errorf("%s has different type", path)
		}
		if strings.HasSuffix(path, ".order") {
			if !reflect.DeepEqual(left, right) {
				return fmt.Errorf("%s has different ID order", path)
			}
		}
	}

	return nil
}
