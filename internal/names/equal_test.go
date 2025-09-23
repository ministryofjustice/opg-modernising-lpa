package names

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type person string

func (p person) FullName() string { return string(p) }

func TestEqual(t *testing.T) {
	testcases := map[string]struct {
		a, b     string
		expected bool
	}{
		"simple":               {"Smith", "Smith", true},
		"simple bad":           {"Smitt", "Smith", false},
		"capitalised":          {"smIth", "SMiTH", true},
		"capitalised bad":      {"smItt", "SMiTH", false},
		"apostrophe":           {"O'Smith", "O‘Smith", true},
		"apostrophe bad":       {"O'Smitt", "O‘Smith", false},
		"other apostrophe":     {"O'Smith", "O’Smith", true},
		"other apostrophe bad": {"O'Smitt", "O’Smith", false},
		"n-dash":               {"Smith–Bloggs", "Smith-Bloggs", true},
		"n-dash bad":           {"Smitt–Bloggs", "Smith-Bloggs", false},
		"m-dash":               {"Smith—Bloggs", "Smith-Bloggs", true},
		"m-dash bad":           {"Smitt—Bloggs", "Smith-Bloggs", false},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Equal(tc.a, tc.b))
		})
	}
}

func TestEqualDoubleBarrel(t *testing.T) {
	testcases := map[string]struct {
		a, b     string
		expected bool
	}{
		"simple":          {"Smith-Bloggs", "Smith", true},
		"simple other":    {"Smith-Bloggs", "Bloggs", true},
		"simple reversed": {"Smith-Bloggs", "Bloggs-Smith", true},
		"simple bad":      {"Smitt-Bloggs", "Smith", false},
		"capitalised":     {"Smith-Bloggs", "BLOGGS", true},
		"apostrophe":      {"O‘Smith-Bloggs", "O'Smith", true},
		"apostrophe bad":  {"O‘Smith-Bloggs", "Smith", false},
		"n-dash":          {"Smith–Bloggs", "Smith", true},
		"m-dash":          {"Smith—Bloggs", "Smith", true},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, EqualDoubleBarrel(tc.a, tc.b))
		})
	}
}

func TestEqualFull(t *testing.T) {
	testcases := map[string]struct {
		a, b     person
		expected bool
	}{
		"simple":               {"Smith", "Smith", true},
		"simple bad":           {"Smitt", "Smith", false},
		"capitalised":          {"smIth", "SMiTH", true},
		"capitalised bad":      {"smItt", "SMiTH", false},
		"apostrophe":           {"O'Smith", "O‘Smith", true},
		"apostrophe bad":       {"O'Smitt", "O‘Smith", false},
		"other apostrophe":     {"O'Smith", "O’Smith", true},
		"other apostrophe bad": {"O'Smitt", "O’Smith", false},
		"n-dash":               {"Smith–Bloggs", "Smith-Bloggs", true},
		"n-dash bad":           {"Smitt–Bloggs", "Smith-Bloggs", false},
		"m-dash":               {"Smith—Bloggs", "Smith-Bloggs", true},
		"m-dash bad":           {"Smitt—Bloggs", "Smith-Bloggs", false},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, EqualFull(tc.a, tc.b))
		})
	}
}
