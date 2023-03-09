package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLanguageFilesMatch(t *testing.T) {
	en := loadTranslations("../lang/en.json")
	cy := loadTranslations("../lang/cy.json")

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
	en := loadTranslations("../lang/en.json")
	cy := loadTranslations("../lang/cy.json")

	for k, v := range en {
		if strings.Contains(v, "'") {
			t.Fail()
			t.Log("lang/en.json:", k)
		}
	}

	for k, v := range cy {
		if strings.Contains(v, "'") {
			t.Fail()
			t.Log("lang/cy.json: ", k)
		}
	}
}

func loadTranslations(path string) map[string]string {
	data, _ := os.ReadFile(path)
	var v map[string]string
	json.Unmarshal(data, &v)

	return v
}
