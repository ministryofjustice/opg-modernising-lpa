package main

import (
	"encoding/json"
	"os"
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
