package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestLanguageFilesMatch(t *testing.T) {
	loadKeys := func(path string) map[string]any {
		data, _ := os.ReadFile(path)
		var v map[string]interface{}
		json.Unmarshal(data, &v)

		return v
	}

	en := loadKeys("lang/en.json")
	cy := loadKeys("lang/cy.json")

	for k := range en {
		if _, ok := cy[k]; !ok {
			t.Fail()
			t.Log("lang/cy.json missing: ", k)
		}
	}

	for k := range cy {
		if _, ok := en[k]; !ok {
			t.Fail()
			t.Log("lang/en.json missing: ", k)
		}
	}
}

func TestApostrophesAreCurly(t *testing.T) {
	loadTranslations := func(path string) map[string]string {
		data, _ := os.ReadFile(path)
		var v map[string]string
		json.Unmarshal(data, &v)

		return v
	}

	en := loadTranslations("lang/en.json")
	cy := loadTranslations("lang/cy.json")

	var failures []string

	for k, v := range en {
		if strings.Contains(v, "'") {
			failures = append(failures, fmt.Sprintf("lang/en.json %s \n", k))
		}
	}

	for k, v := range cy {
		if strings.Contains(v, "'") {
			failures = append(failures, fmt.Sprintf("lang/cy.json %s \n", k))
		}
	}

	if len(failures) > 0 {
		t.Log("non-curly apostrophe present in translations:\n", failures)
		t.Fail()
	}
}
