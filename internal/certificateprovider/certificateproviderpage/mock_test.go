package certificateproviderpage

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

var (
	testAddress = place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "e",
	}
	expectedError = errors.New("err")
	testAppData   = appcontext.Data{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Localizer: fakeLocalizer{},
	}
	testNow   = time.Date(2020, time.February, 3, 12, 13, 14, 15, time.UTC)
	testNowFn = func() time.Time { return testNow }
)

type fakeLocalizer struct{}

func (f fakeLocalizer) Concat(list []string, joiner string) string { return "" }
func (f fakeLocalizer) Count(messageID string, count int) string   { return "" }
func (f fakeLocalizer) Format(messageID string, data map[string]any) string {
	var s strings.Builder
	s.WriteString(messageID)
	for _, k := range slices.Sorted(maps.Keys(data)) {
		s.WriteByte(':')
		s.WriteString(k)
		s.WriteByte('=')
		fmt.Fprint(&s, data[k])
	}

	return s.String()
}
func (f fakeLocalizer) FormatCount(messageID string, count int, data map[string]any) string {
	return ""
}
func (f fakeLocalizer) FormatDate(t date.TimeOrDate) string { return t.Format(time.DateOnly) }
func (f fakeLocalizer) FormatDateTime(t time.Time) string   { return t.Format(time.RFC3339) }
func (f fakeLocalizer) FormatTime(t time.Time) string       { return t.Format(time.TimeOnly) }
func (f fakeLocalizer) Lang() localize.Lang                 { return localize.En }
func (f fakeLocalizer) Possessive(s string) string          { return s + "'s" }
func (f fakeLocalizer) ShowTranslationKeys() bool           { return false }
func (f fakeLocalizer) SetShowTranslationKeys(s bool)       {}
func (f fakeLocalizer) T(s string) string                   { return s }
