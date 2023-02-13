package templatefn

import (
	"html/template"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestIsEnglish(t *testing.T) {
	assert.True(t, isEnglish(localize.En))
	assert.False(t, isEnglish(localize.Cy))
}

func TestIsWelsh(t *testing.T) {
	assert.True(t, isWelsh(localize.Cy))
	assert.False(t, isWelsh(localize.En))
}

func TestInput(t *testing.T) {
	top := 1
	name := "name"
	label := "label"
	value := 2

	v := input(top, name, label, value)

	assert.Equal(t, top, v["top"])
	assert.Equal(t, name, v["name"])
	assert.Equal(t, label, v["label"])
	assert.Equal(t, value, v["value"])
}

func TestInputWithAttrs(t *testing.T) {
	top := 1
	name := "name"
	label := "label"
	value := 2
	hint := "this"

	v := input(top, name, label, value, "hint", hint)

	assert.Equal(t, top, v["top"])
	assert.Equal(t, name, v["name"])
	assert.Equal(t, label, v["label"])
	assert.Equal(t, value, v["value"])
	assert.Equal(t, hint, v["hint"])
}

func TestInputWithUnevenAttrs(t *testing.T) {
	assert.Panics(t, func() { input(1, "name", "label", 2, "hint") })
}

func TestItems(t *testing.T) {
	top := 1
	name := "name"
	value := 2
	items1 := "1"
	items2 := "2"

	v := items(top, name, value, items1, items2)

	assert.Equal(t, top, v["top"])
	assert.Equal(t, name, v["name"])
	assert.Equal(t, value, v["value"])
	assert.Equal(t, []interface{}{items1, items2}, v["items"])
}

func TestItem(t *testing.T) {
	value := "value"
	label := "label"

	v := item(value, label)

	assert.Equal(t, value, v["value"])
	assert.Equal(t, label, v["label"])
}

func TestItemWithAttrs(t *testing.T) {
	value := "value"
	label := "label"
	hint := "this"

	v := item(value, label, "hint", hint)

	assert.Equal(t, value, v["value"])
	assert.Equal(t, label, v["label"])
	assert.Equal(t, hint, v["hint"])
}

func TestItemWithUnevenAttrs(t *testing.T) {
	assert.Panics(t, func() { item("value", "label", "hint") })
}

func TestFieldID(t *testing.T) {
	assert.Equal(t, "field", fieldID("field", 0))
	assert.Equal(t, "field-2", fieldID("field", 1))
	assert.Equal(t, "field-3", fieldID("field", 2))
}

func TestErrorMessage(t *testing.T) {
	top := 1
	name := "name"

	v := errorMessage(top, name)

	assert.Equal(t, top, v["top"])
	assert.Equal(t, name, v["name"])
}

func TestDetails(t *testing.T) {
	top := 1
	name := "name"
	detail := "detail"
	open := true

	v := details(top, name, detail, open)

	assert.Equal(t, top, v["top"])
	assert.Equal(t, name, v["name"])
	assert.Equal(t, detail, v["detail"])
	assert.Equal(t, open, v["open"])
}

func TestInc(t *testing.T) {
	assert.Equal(t, 1, inc(0))
	assert.Equal(t, 2, inc(1))
}

func TestLink(t *testing.T) {
	assert.Equal(t, "/dashboard", link(page.AppData{}, "/dashboard"))
	assert.Equal(t, "/cy/dashboard", link(page.AppData{Lang: localize.Cy}, "/dashboard"))
	assert.Equal(t, "/lpa/123/somewhere", link(page.AppData{LpaID: "123"}, "/somewhere"))
	assert.Equal(t, "/cy/lpa/123/somewhere", link(page.AppData{Lang: localize.Cy, LpaID: "123"}, "/somewhere"))
}

func TestContains(t *testing.T) {
	assert.True(t, contains("b", []string{"a", "b", "c"}))
	assert.False(t, contains("d", []string{"a", "b", "c"}))
}

func TestTr(t *testing.T) {
	app := page.AppData{
		Localizer: localize.NewBundle("testdata/en.json").For("en"),
	}

	assert.Equal(t, "hi", tr(app, "message-id"))
	assert.Equal(t, "", tr(app, ""))
}

func TestTrFormat(t *testing.T) {
	app := page.AppData{
		Localizer: localize.NewBundle("testdata/en.json").For("en"),
	}

	assert.Equal(t, "hi Person", trFormat(app, "with-format", "name", "Person"))
	assert.Equal(t, "", trFormat(app, "", "name", "Person"))
}

func TestTrHtml(t *testing.T) {
	app := page.AppData{
		Localizer: localize.NewBundle("testdata/en.json").For("en"),
	}

	assert.Equal(t, template.HTML("hi"), trHtml(app, "message-id"))
	assert.Equal(t, template.HTML(""), trHtml(app, ""))
}

func TestTrFormatHtml(t *testing.T) {
	app := page.AppData{
		Localizer: localize.NewBundle("testdata/en.json").For("en"),
	}

	assert.Equal(t, template.HTML("hi Person"), trFormatHtml(app, "with-format", "name", "Person"))
	assert.Equal(t, template.HTML(""), trFormatHtml(app, "", "name", "Person"))
}

func TestTrCount(t *testing.T) {
	enApp := page.AppData{
		Localizer: localize.NewBundle("testdata/en.json").For("en"),
	}

	assert.Equal(t, "hi one", trCount(enApp, "with-count", 1))
	assert.Equal(t, "hi other", trCount(enApp, "with-count", 2))
	assert.Equal(t, "", trCount(enApp, "", 2))

	cyApp := page.AppData{
		Localizer: localize.NewBundle("testdata/cy.json").For("cy"),
	}

	assert.Equal(t, "cy one", trCount(cyApp, "with-count", 1))
	assert.Equal(t, "cy two", trCount(cyApp, "with-count", 2))
	assert.Equal(t, "cy few", trCount(cyApp, "with-count", 3))
	assert.Equal(t, "cy other", trCount(cyApp, "with-count", 4))
	assert.Equal(t, "cy other", trCount(cyApp, "with-count", 5))
	assert.Equal(t, "cy many", trCount(cyApp, "with-count", 6))
	assert.Equal(t, "cy other", trCount(cyApp, "with-count", 7))
	assert.Equal(t, "", trCount(cyApp, "", 7))
}

func TestTrFormatCount(t *testing.T) {
	enApp := page.AppData{
		Localizer: localize.NewBundle("testdata/en.json").For("en"),
	}

	assert.Equal(t, "hi 1 one Person", trFormatCount(enApp, "with-format-count", 1, "name", "Person"))
	assert.Equal(t, "hi 2 other Person", trFormatCount(enApp, "with-format-count", 2, "name", "Person"))
	assert.Equal(t, "", trFormatCount(enApp, "", 2, "name", "Person"))

	cyApp := page.AppData{
		Localizer: localize.NewBundle("testdata/cy.json").For("cy"),
	}

	assert.Equal(t, "cy hi 1 one Person", trFormatCount(cyApp, "with-format-count", 1, "name", "Person"))
	assert.Equal(t, "cy hi 2 two Person", trFormatCount(cyApp, "with-format-count", 2, "name", "Person"))
	assert.Equal(t, "cy hi 3 few Person", trFormatCount(cyApp, "with-format-count", 3, "name", "Person"))
	assert.Equal(t, "cy hi 4 other Person", trFormatCount(cyApp, "with-format-count", 4, "name", "Person"))
	assert.Equal(t, "cy hi 5 other Person", trFormatCount(cyApp, "with-format-count", 5, "name", "Person"))
	assert.Equal(t, "cy hi 6 many Person", trFormatCount(cyApp, "with-format-count", 6, "name", "Person"))
	assert.Equal(t, "cy hi 7 other Person", trFormatCount(cyApp, "with-format-count", 7, "name", "Person"))
	assert.Equal(t, "", trFormatCount(cyApp, "", 7, "name", "Person"))
}

func TestNow(t *testing.T) {
	assert.WithinDuration(t, time.Now(), now(), time.Millisecond)
}

func TestAddDays(t *testing.T) {
	assert.Equal(t, time.Date(2020, time.January, 7, 3, 4, 5, 6, time.UTC), addDays(5, time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC)))
}

func TestFormatDate(t *testing.T) {
	assert.Equal(t, "7 March 2020", formatDate(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 March 2020", formatDate(date.New("2020", "3", "7")))
}

func TestFormatDateTime(t *testing.T) {
	assert.Equal(t, "7 March 2020 at 03:04", formatDateTime(time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
}

func TestLowerFirst(t *testing.T) {
	assert.Equal(t, "hELLO", lowerFirst("HELLO"))
	assert.Equal(t, "hello", lowerFirst("hello"))
}

func TestListAttorneysWithAttorneys(t *testing.T) {
	attorneys := actor.Attorneys{
		{ID: "123"},
		{ID: "123"},
	}

	app := page.AppData{SessionID: "abc", Page: "/here"}
	withHeaders := true
	lpa := &page.Lpa{}
	attorneyType := "attorney"

	want := map[string]interface{}{
		"Attorneys":    attorneys,
		"App":          app,
		"WithHeaders":  withHeaders,
		"Lpa":          lpa,
		"AttorneyType": attorneyType,
		"DetailsPath":  app.Paths.ChooseAttorneys + "?from=/here",
		"AddressPath":  app.Paths.ChooseAttorneysAddress + "?from=/here",
		"RemovePath":   app.Paths.RemoveAttorney + "?from=/here",
	}

	got := listAttorneys(attorneys, app, attorneyType, withHeaders, lpa)

	assert.Equal(t, want, got)
}

func TestListAttorneysWithReplacementAttorneys(t *testing.T) {
	attorneys := actor.Attorneys{
		{ID: "123"},
		{ID: "123"},
	}

	app := page.AppData{SessionID: "abc", Page: "/here"}
	withHeaders := true
	lpa := &page.Lpa{}
	attorneyType := "replacement"

	want := map[string]interface{}{
		"Attorneys":    attorneys,
		"App":          app,
		"WithHeaders":  withHeaders,
		"Lpa":          lpa,
		"AttorneyType": attorneyType,
		"DetailsPath":  app.Paths.ChooseReplacementAttorneys + "?from=/here",
		"AddressPath":  app.Paths.ChooseReplacementAttorneysAddress + "?from=/here",
		"RemovePath":   app.Paths.RemoveReplacementAttorney + "?from=/here",
	}

	got := listAttorneys(attorneys, app, attorneyType, withHeaders, lpa)

	assert.Equal(t, want, got)
}

func TestWarning(t *testing.T) {
	app := page.AppData{SessionID: "abc"}
	content := "content"

	v := warning(app, content)

	assert.Equal(t, app, v["app"])
	assert.Equal(t, content, v["content"])
}

func TestListPeopleToNotify(t *testing.T) {
	app := page.AppData{SessionID: "abc"}
	withHeaders := true
	lpa := &page.Lpa{}

	want := map[string]interface{}{
		"App":         app,
		"WithHeaders": withHeaders,
		"Lpa":         lpa,
	}

	got := listPeopleToNotify(app, withHeaders, lpa)

	assert.Equal(t, want, got)
}

func TestProgressbar(t *testing.T) {
	app := page.AppData{SessionID: "abc"}
	lpa := &page.Lpa{}

	want := map[string]interface{}{
		"App": app,
		"Lpa": lpa,
	}

	got := progressBar(app, lpa)

	assert.Equal(t, want, got)
}

func TestListPeopleNamedOnLpa(t *testing.T) {
	app := page.AppData{SessionID: "abc"}
	withHeaders := true
	lpa := &page.Lpa{}

	want := map[string]interface{}{
		"App":               app,
		"ShowPeopleHeaders": withHeaders,
		"Lpa":               lpa,
	}

	got := peopleNamedOnLpa(app, lpa, withHeaders)

	assert.Equal(t, want, got)
}
