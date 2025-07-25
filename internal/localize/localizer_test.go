package localize

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/stretchr/testify/assert"
)

func TestShowTranslationKey(t *testing.T) {
	localizer := defaultLocalizer{showTranslationKeys: true}
	assert.True(t, localizer.ShowTranslationKeys())
}

func TestPossessive(t *testing.T) {
	en := defaultLocalizer{lang: En}
	cy := defaultLocalizer{lang: Cy}

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
	en := defaultLocalizer{lang: En}
	cy := defaultLocalizer{lang: Cy}

	assert.Equal(t, "7 March 2020", en.FormatDate(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 March 2020", en.FormatDate(date.New("2020", "3", "7")))

	assert.Equal(t, "7 Mawrth 2020", cy.FormatDate(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020", cy.FormatDate(date.New("2020", "3", "7")))
}

func TestFormatTime(t *testing.T) {
	en := defaultLocalizer{lang: En}
	cy := defaultLocalizer{lang: Cy}

	assert.Equal(t, "", en.FormatTime(time.Time{}))

	assert.Equal(t, "3:04am", en.FormatTime(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "4:04am", en.FormatTime(time.Date(2020, time.March, 30, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "3:04yb", cy.FormatTime(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "4:04yb", cy.FormatTime(time.Date(2020, time.March, 30, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "3:04pm", en.FormatTime(time.Date(2020, time.March, 7, 15, 4, 5, 6, time.UTC)))
	assert.Equal(t, "3:04yp", cy.FormatTime(time.Date(2020, time.March, 7, 15, 4, 5, 6, time.UTC)))

	assert.Equal(t, "12:00am", en.FormatTime(time.Date(2020, time.March, 7, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, "12:00yb", cy.FormatTime(time.Date(2020, time.March, 7, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, "12:00pm", en.FormatTime(time.Date(2020, time.March, 7, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, "12:00yp", cy.FormatTime(time.Date(2020, time.March, 7, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, "12:00am", en.FormatTime(time.Date(2020, time.March, 7, 24, 0, 0, 0, time.UTC)))
	assert.Equal(t, "12:00yb", cy.FormatTime(time.Date(2020, time.March, 7, 24, 0, 0, 0, time.UTC)))
}

func TestFormatDateTime(t *testing.T) {
	en := defaultLocalizer{lang: En}
	cy := defaultLocalizer{lang: Cy}

	assert.Equal(t, "", en.FormatDateTime(time.Time{}))

	assert.Equal(t, "3:04am on 7 March 2020", en.FormatDateTime(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "4:04am on 30 March 2020", en.FormatDateTime(time.Date(2020, time.March, 30, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020 am 3:04yb", cy.FormatDateTime(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "30 Mawrth 2020 am 4:04yb", cy.FormatDateTime(time.Date(2020, time.March, 30, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "3:04pm on 7 March 2020", en.FormatDateTime(time.Date(2020, time.March, 7, 15, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020 am 3:04yp", cy.FormatDateTime(time.Date(2020, time.March, 7, 15, 4, 5, 6, time.UTC)))

	assert.Equal(t, "12:00am on 7 March 2020", en.FormatDateTime(time.Date(2020, time.March, 7, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020 am 12:00yb", cy.FormatDateTime(time.Date(2020, time.March, 7, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, "12:00pm on 7 March 2020", en.FormatDateTime(time.Date(2020, time.March, 7, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020 am 12:00yp", cy.FormatDateTime(time.Date(2020, time.March, 7, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, "12:00am on 8 March 2020", en.FormatDateTime(time.Date(2020, time.March, 7, 24, 0, 0, 0, time.UTC)))
	assert.Equal(t, "8 Mawrth 2020 am 12:00yb", cy.FormatDateTime(time.Date(2020, time.March, 7, 24, 0, 0, 0, time.UTC)))
}

func TestLowerFirst(t *testing.T) {
	assert.Equal(t, "hELLO", LowerFirst("HELLO"))
	assert.Equal(t, "hello", LowerFirst("hello"))
}
