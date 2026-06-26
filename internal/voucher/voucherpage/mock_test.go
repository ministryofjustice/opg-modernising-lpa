package voucherpage

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
)

var (
	expectedError = errors.New("err")
	testAppData   = appcontext.Data{LpaID: "lpa-id", Localizer: fakeLocalizer{}}
	testNow       = time.Now()
	testNowFn     = func() time.Time { return testNow }
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
func (f fakeLocalizer) FormatDate(t date.TimeOrDate) string { return "" }
func (f fakeLocalizer) FormatDateTime(t time.Time) string   { return "" }
func (f fakeLocalizer) FormatTime(t time.Time) string       { return "" }
func (f fakeLocalizer) Lang() localize.Lang                 { return localize.En }
func (f fakeLocalizer) Possessive(s string) string          { return "" }
func (f fakeLocalizer) ShowTranslationKeys() bool           { return false }
func (f fakeLocalizer) SetShowTranslationKeys(s bool)       {}
func (f fakeLocalizer) T(s string) string                   { return s }
