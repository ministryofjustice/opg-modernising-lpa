package localize

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/stretchr/testify/assert"
)

func TestNewBundle(t *testing.T) {
	assert := assert.New(t)
	bundle, err := NewBundle("testdata/en.json", "testdata/cy.json")
	assert.Nil(err)

	en := bundle.For(En)
	assert.Equal("A", en.T("a"))
	assert.Equal("key does not exist", en.T("key does not exist"))

	assert.Equal("A person", en.Format("af", map[string]interface{}{"x": "person"}))

	assert.Equal("1 ONE", en.Count("c", 1))
	assert.Equal("2 OTHER", en.Count("c", 2))
	assert.Equal("key does not exist", en.Count("key does not exist", 3))

	assert.Equal("1 ONE FORMATTED", en.FormatCount("d", 1, map[string]interface{}{"x": "FORMATTED"}))
	assert.Equal("2 OTHER FORMATTED", en.FormatCount("d", 2, map[string]interface{}{"x": "FORMATTED"}))

	cy := bundle.For(Cy)
	assert.Equal("C", cy.T("a"))

	assert.Equal("C person", cy.Format("af", map[string]interface{}{"x": "person"}))

	assert.Equal("1 one", cy.Count("c", 1))
	assert.Equal("2 two", cy.Count("c", 2))
	assert.Equal("3 few", cy.Count("c", 3))
	assert.Equal("4 other", cy.Count("c", 4))
	assert.Equal("5 other", cy.Count("c", 5))
	assert.Equal("6 many", cy.Count("c", 6))
	assert.Equal("7 other", cy.Count("c", 7))

	assert.Equal("1 one formatted", cy.FormatCount("d", 1, map[string]interface{}{"x": "formatted"}))
	assert.Equal("2 two formatted", cy.FormatCount("d", 2, map[string]interface{}{"x": "formatted"}))
	assert.Equal("3 few formatted", cy.FormatCount("d", 3, map[string]interface{}{"x": "formatted"}))
	assert.Equal("4 other formatted", cy.FormatCount("d", 4, map[string]interface{}{"x": "formatted"}))
	assert.Equal("5 other formatted", cy.FormatCount("d", 5, map[string]interface{}{"x": "formatted"}))
	assert.Equal("6 many formatted", cy.FormatCount("d", 6, map[string]interface{}{"x": "formatted"}))
	assert.Equal("7 other formatted", cy.FormatCount("d", 7, map[string]interface{}{"x": "formatted"}))
}

func TestNewBundleWhenBadFormat(t *testing.T) {
	_, err := NewBundle("testdata/bad/en.json")
	assert.NotNil(t, err)

	_, err = NewBundle("testdata/bad/cy.json")
	assert.NotNil(t, err)
}

func TestNewBundleWhenMissingFile(t *testing.T) {
	_, err := NewBundle("testdata/a.json")
	assert.NotNil(t, err)
}

func TestNewBundleWhenOtherLang(t *testing.T) {
	_, err := NewBundle("testdata/zz.json")
	assert.NotNil(t, err)
}

func TestNewBundleWithTransKeys(t *testing.T) {
	assert := assert.New(t)
	bundle, _ := NewBundle("testdata/en.json", "testdata/cy.json")

	en := bundle.For(En)
	en.SetShowTranslationKeys(true)

	assert.Equal("{A} [a]", en.T("a"))
	assert.Equal("{key does not exist} [key does not exist]", en.T("key does not exist"))

	assert.Equal("{A person} [af]", en.Format("af", map[string]interface{}{"x": "person"}))

	assert.Equal("{1 ONE} [c]", en.Count("c", 1))
	assert.Equal("{2 OTHER} [c]", en.Count("c", 2))

	assert.Equal("{1 ONE FORMATTED} [d]", en.FormatCount("d", 1, map[string]interface{}{"x": "FORMATTED"}))
	assert.Equal("{2 OTHER FORMATTED} [d]", en.FormatCount("d", 2, map[string]interface{}{"x": "FORMATTED"}))

	cy := bundle.For(Cy)
	cy.SetShowTranslationKeys(true)

	assert.Equal("{C} [a]", cy.T("a"))

	assert.Equal("{C person} [af]", cy.Format("af", map[string]interface{}{"x": "person"}))

	assert.Equal("{1 one} [c]", cy.Count("c", 1))
	assert.Equal("{2 two} [c]", cy.Count("c", 2))
	assert.Equal("{3 few} [c]", cy.Count("c", 3))
	assert.Equal("{4 other} [c]", cy.Count("c", 4))
	assert.Equal("{5 other} [c]", cy.Count("c", 5))
	assert.Equal("{6 many} [c]", cy.Count("c", 6))
	assert.Equal("{7 other} [c]", cy.Count("c", 7))

	assert.Equal("{1 one formatted} [d]", cy.FormatCount("d", 1, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{2 two formatted} [d]", cy.FormatCount("d", 2, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{3 few formatted} [d]", cy.FormatCount("d", 3, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{4 other formatted} [d]", cy.FormatCount("d", 4, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{5 other formatted} [d]", cy.FormatCount("d", 5, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{6 many formatted} [d]", cy.FormatCount("d", 6, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{7 other formatted} [d]", cy.FormatCount("d", 7, map[string]interface{}{"x": "formatted"}))
}

func TestShowTranslationKey(t *testing.T) {
	localizer := Localizer{showTranslationKeys: true}
	assert.True(t, localizer.ShowTranslationKeys())
}

func TestPossessive(t *testing.T) {
	en := Localizer{Lang: En}
	cy := Localizer{Lang: Cy}

	testCases := map[string]struct {
		Str      string
		Lang     Lang
		Expected string
	}{
		"En - not ending in s": {
			Str:      "a",
			Expected: "a’s",
		},
		"En - ending in s": {
			Str:      "s",
			Expected: "s’",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, en.Possessive(tc.Str))
			assert.Equal(t, tc.Str, cy.Possessive(tc.Str))
		})
	}
}

func TestConcat(t *testing.T) {
	bundle, _ := NewBundle("testdata/en.json", "testdata/cy.json")
	en := bundle.For(En)

	assert.Equal(t, "Bob Smith, Alice Jones, John Doe or Paul Compton", en.Concat([]string{"Bob Smith", "Alice Jones", "John Doe", "Paul Compton"}, "or"))
	assert.Equal(t, "Bob Smith, Alice Jones and John Doe", en.Concat([]string{"Bob Smith", "Alice Jones", "John Doe"}, "and"))
	assert.Equal(t, "Bob Smith and John Doe", en.Concat([]string{"Bob Smith", "John Doe"}, "and"))
	assert.Equal(t, "Bob Smith", en.Concat([]string{"Bob Smith"}, "and"))
	assert.Equal(t, "", en.Concat([]string{}, "and"))

	cy := bundle.For(Cy)
	assert.Equal(t, "Bob Smith, Alice Jones, John Doe neu Paul Compton", cy.Concat([]string{"Bob Smith", "Alice Jones", "John Doe", "Paul Compton"}, "or"))
	assert.Equal(t, "Bob Smith, Alice Jones a John Doe", cy.Concat([]string{"Bob Smith", "Alice Jones", "John Doe"}, "and"))
	assert.Equal(t, "Bob Smith a John Doe", cy.Concat([]string{"Bob Smith", "John Doe"}, "and"))
	assert.Equal(t, "Bob Smith", cy.Concat([]string{"Bob Smith"}, "and"))
	assert.Equal(t, "", cy.Concat([]string{}, "and"))
}

func TestFormatDate(t *testing.T) {
	en := Localizer{Lang: En}
	cy := Localizer{Lang: Cy}

	assert.Equal(t, "7 March 2020", en.FormatDate(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 March 2020", en.FormatDate(date.New("2020", "3", "7")))

	assert.Equal(t, "7 Mawrth 2020", cy.FormatDate(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020", cy.FormatDate(date.New("2020", "3", "7")))
}

func TestFormatDateTime(t *testing.T) {
	en := Localizer{Lang: En}
	cy := Localizer{Lang: Cy}

	assert.Equal(t, "7 March 2020 at 3:04am", en.FormatDateTime(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020 am 3:04yb", cy.FormatDateTime(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 March 2020 at 3:04pm", en.FormatDateTime(time.Date(2020, time.March, 7, 15, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020 am 3:04yp", cy.FormatDateTime(time.Date(2020, time.March, 7, 15, 4, 5, 6, time.UTC)))

	assert.Equal(t, "7 March 2020 at 12:00am", en.FormatDateTime(time.Date(2020, time.March, 7, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020 am 12:00yb", cy.FormatDateTime(time.Date(2020, time.March, 7, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, "7 March 2020 at 12:00pm", en.FormatDateTime(time.Date(2020, time.March, 7, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020 am 12:00yp", cy.FormatDateTime(time.Date(2020, time.March, 7, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, "8 March 2020 at 12:00am", en.FormatDateTime(time.Date(2020, time.March, 7, 24, 0, 0, 0, time.UTC)))
	assert.Equal(t, "8 Mawrth 2020 am 12:00yb", cy.FormatDateTime(time.Date(2020, time.March, 7, 24, 0, 0, 0, time.UTC)))
}

func TestLowerFirst(t *testing.T) {
	assert.Equal(t, "hELLO", LowerFirst("HELLO"))
	assert.Equal(t, "hello", LowerFirst("hello"))
}
