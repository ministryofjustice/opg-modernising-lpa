package templatefn

import (
	"fmt"
	"html/template"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	globals := &Globals{Tag: "abc"}

	fns := All(globals)
	assert.Equal(t, globals, fns["global"].(func() *Globals)())
}

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

func TestButton(t *testing.T) {
	appData := page.AppData{Path: "1"}

	assert.Equal(t, map[string]any{"app": appData, "label": "label", "link": "a"}, button(appData, "label", "link", "a"))
}

func TestButtonWithUnevenAttrs(t *testing.T) {
	appData := page.AppData{Path: "1"}

	assert.Panics(t, func() { button(appData, "label", "a") })
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
}

func TestFromLink(t *testing.T) {
	assert.Equal(t, "/lpa/lpa-id/your-details?from=/previous#f-first-names",
		fromLink(page.AppData{LpaID: "lpa-id", Page: "/previous"}, page.Paths.YourDetails, "#f-first-names"))
	assert.Equal(t, "/cy/attorney/lpa-id/confirm-your-details?from=/previous",
		fromLink(page.AppData{LpaID: "lpa-id", Page: "/previous", Lang: localize.Cy}, page.Paths.Attorney.ConfirmYourDetails, ""))
}

func TestFromLinkActor(t *testing.T) {
	uid := actoruid.New()
	assert.Equal(t, fmt.Sprintf("/lpa/lpa-id/your-details?from=/previous&id=%s#f-first-names", uid.String()),
		fromLinkActor(page.AppData{LpaID: "lpa-id", Page: "/previous"}, page.Paths.YourDetails, uid, "#f-first-names"))
	assert.Equal(t, "/cy/attorney/lpa-id/confirm-your-details?from=/previous&id="+uid.String(),
		fromLinkActor(page.AppData{LpaID: "lpa-id", Page: "/previous", Lang: localize.Cy}, page.Paths.Attorney.ConfirmYourDetails, uid, ""))
}

func TestCheckboxEq(t *testing.T) {
	assert.True(t, checkboxEq("b", []string{"a", "b", "c"}))
	assert.False(t, checkboxEq("d", []string{"a", "b", "c"}))

	assert.True(t, checkboxEq("b", "b"))
	assert.False(t, checkboxEq("b", "d"))

	assert.False(t, checkboxEq("", nil))
}

func TestTr(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json")
	app := page.AppData{
		Localizer: bundle.For(localize.En),
	}

	assert.Equal(t, "hi", tr(app, "message-id"))
	assert.Equal(t, "", tr(app, ""))
}

func TestTrFormat(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json")
	app := page.AppData{
		Localizer: bundle.For(localize.En),
	}

	assert.Equal(t, "hi Person", trFormat(app, "with-format", "name", "Person"))
	assert.Equal(t, "", trFormat(app, "", "name", "Person"))
}

func TestTrHtml(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json")
	app := page.AppData{
		Localizer: bundle.For(localize.En),
	}

	assert.Equal(t, template.HTML("hi"), trHtml(app, "message-id"))
	assert.Equal(t, template.HTML(""), trHtml(app, ""))
}

func TestTrFormatHtml(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json")
	app := page.AppData{
		Localizer: bundle.For(localize.En),
	}

	assert.Equal(t, template.HTML("hi Person"), trFormatHtml(app, "with-format", "name", "Person"))
	assert.Equal(t, template.HTML(""), trFormatHtml(app, "", "name", "Person"))

	assert.Equal(t, template.HTML("hi &lt;script&gt;alert(&#39;hi&#39;);&lt;/script&gt;"), trFormatHtml(app, "with-format", "name", "<script>alert('hi');</script>"))
}

func TestTrCount(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json", "testdata/cy.json")

	enApp := page.AppData{Localizer: bundle.For(localize.En)}
	assert.Equal(t, "hi one", trCount(enApp, "with-count", 1))
	assert.Equal(t, "hi other", trCount(enApp, "with-count", 2))
	assert.Equal(t, "", trCount(enApp, "", 2))

	cyApp := page.AppData{Localizer: bundle.For(localize.Cy)}
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
	bundle, _ := localize.NewBundle("testdata/en.json")
	enApp := page.AppData{
		Localizer: bundle.For(localize.En),
	}

	assert.Equal(t, "hi 1 one Person", trFormatCount(enApp, "with-format-count", 1, "name", "Person"))
	assert.Equal(t, "hi 2 other Person", trFormatCount(enApp, "with-format-count", 2, "name", "Person"))
	assert.Equal(t, "", trFormatCount(enApp, "", 2, "name", "Person"))

	bundle, _ = localize.NewBundle("testdata/cy.json")
	cyApp := page.AppData{
		Localizer: bundle.For(localize.Cy),
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
	bundle, _ := localize.NewBundle("testdata/en.json", "testdata/cy.json")
	appEn := page.AppData{Localizer: bundle.For(localize.En)}
	appCy := page.AppData{Localizer: bundle.For(localize.Cy)}

	assert.Equal(t, "7 March 2020", formatDate(appEn, time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 March 2020", formatDate(appEn, date.New("2020", "3", "7")))

	assert.Equal(t, "7 Mawrth 2020", formatDate(appCy, time.Date(2020, time.March, 7, 3, 4, 5, 6, time.UTC)))
	assert.Equal(t, "7 Mawrth 2020", formatDate(appCy, date.New("2020", "3", "7")))
}

func TestFormatTime(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json", "testdata/cy.json")
	appEn := page.AppData{Localizer: bundle.For(localize.En)}
	appCy := page.AppData{Localizer: bundle.For(localize.Cy)}

	assert.Equal(t, "3:04am", formatTime(appEn, time.Date(2020, time.March, 7, 3, 4, 0, 0, time.UTC)))

	assert.Equal(t, "3:04yb", formatTime(appCy, time.Date(2020, time.March, 7, 3, 4, 0, 0, time.UTC)))
}

func TestFormatDateTime(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json", "testdata/cy.json")
	appEn := page.AppData{Localizer: bundle.For(localize.En)}
	appCy := page.AppData{Localizer: bundle.For(localize.Cy)}

	assert.Equal(t, "7 March 2020 at 3:04am", formatDateTime(appEn, time.Date(2020, time.March, 7, 3, 4, 0, 0, time.UTC)))

	assert.Equal(t, "7 Mawrth 2020 am 3:04yb", formatDateTime(appCy, time.Date(2020, time.March, 7, 3, 4, 0, 0, time.UTC)))
}

func TestFormatPhone(t *testing.T) {
	assert.Equal(t, "07123 456 789", formatPhone("07 12 34 56 78 9"))
	assert.Equal(t, "+44 7123 456 789", formatPhone("+44 71 23 45 67 89"))
	assert.Equal(t, "+44 something else", formatPhone("+44 something else"))
}

func TestListAttorneysWithAttorneys(t *testing.T) {
	app := page.AppData{SessionID: "abc", Page: "/here", ActorType: actor.TypeDonor, LpaID: "lpa-id"}
	headingLevel := 3
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	attorneyLinks := attorneySummaryDataLinks{
		Attorney:                page.Paths.ChooseAttorneys.Format("lpa-id") + "?from=/here",
		AttorneyAddress:         page.Paths.ChooseAttorneysAddress.Format("lpa-id") + "?from=/here",
		RemoveAttorney:          page.Paths.RemoveAttorney.Format("lpa-id") + "?from=/here",
		TrustCorporation:        page.Paths.EnterTrustCorporation.Format("lpa-id") + "?from=/here",
		TrustCorporationAddress: page.Paths.EnterTrustCorporationAddress.Format("lpa-id") + "?from=/here",
		RemoveTrustCorporation:  page.Paths.RemoveTrustCorporation.Format("lpa-id") + "?from=/here",
	}

	replacementLinks := attorneySummaryDataLinks{
		Attorney:                page.Paths.ChooseReplacementAttorneys.Format("lpa-id") + "?from=/here",
		AttorneyAddress:         page.Paths.ChooseReplacementAttorneysAddress.Format("lpa-id") + "?from=/here",
		RemoveAttorney:          page.Paths.RemoveReplacementAttorney.Format("lpa-id") + "?from=/here",
		TrustCorporation:        page.Paths.EnterReplacementTrustCorporation.Format("lpa-id") + "?from=/here",
		TrustCorporationAddress: page.Paths.EnterReplacementTrustCorporationAddress.Format("lpa-id") + "?from=/here",
		RemoveTrustCorporation:  page.Paths.RemoveReplacementTrustCorporation.Format("lpa-id") + "?from=/here",
	}

	lpaStoreAttorneys := []lpastore.Attorney{
		{UID: uid1},
		{UID: uid2},
	}
	lpaStoreTrustCorporation := lpastore.TrustCorporation{Name: "a"}

	actorAttorneys := []donordata.Attorney{
		{UID: uid1},
		{UID: uid2},
	}
	actorTrustCorporation := donordata.TrustCorporation{Name: "a"}

	testcases := map[string]struct {
		attorneys    any
		data         attorneySummaryData
		attorneyType string
	}{
		"lpastore": {
			attorneys: lpastore.Attorneys{
				Attorneys:        lpaStoreAttorneys,
				TrustCorporation: lpaStoreTrustCorporation,
			},
			attorneyType: "attorney",
			data: attorneySummaryData{
				TrustCorporation: lpaStoreTrustCorporation,
				Attorneys:        lpaStoreAttorneys,
				App:              app,
				HeadingLevel:     headingLevel,
				CanChange:        true,
				Link:             attorneyLinks,
			},
		},
		"dynamo": {
			attorneys: donordata.Attorneys{
				Attorneys:        actorAttorneys,
				TrustCorporation: actorTrustCorporation,
			},
			attorneyType: "attorney",
			data: attorneySummaryData{
				TrustCorporation: lpaStoreTrustCorporation,
				Attorneys:        lpaStoreAttorneys,
				App:              app,
				HeadingLevel:     headingLevel,
				CanChange:        true,
				Link:             attorneyLinks,
			},
		},
		"lpastore replacement": {
			attorneys: lpastore.Attorneys{
				Attorneys:        lpaStoreAttorneys,
				TrustCorporation: lpaStoreTrustCorporation,
			},
			attorneyType: "replacement",
			data: attorneySummaryData{
				TrustCorporation: lpaStoreTrustCorporation,
				Attorneys:        lpaStoreAttorneys,
				App:              app,
				HeadingLevel:     headingLevel,
				CanChange:        true,
				Link:             replacementLinks,
			},
		},
		"dynamo replacement": {
			attorneys: donordata.Attorneys{
				Attorneys:        actorAttorneys,
				TrustCorporation: actorTrustCorporation,
			},
			attorneyType: "replacement",
			data: attorneySummaryData{
				TrustCorporation: lpaStoreTrustCorporation,
				Attorneys:        lpaStoreAttorneys,
				App:              app,
				HeadingLevel:     headingLevel,
				CanChange:        true,
				Link:             replacementLinks,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			got := listAttorneys(app, tc.attorneys, tc.attorneyType, headingLevel, true)

			assert.Equal(t, tc.data, got)
		})
	}
}

func TestListAttorneysWithIncorrectType(t *testing.T) {
	assert.Panics(t, func() {
		listAttorneys(page.AppData{}, 5, "attorney", 3, true)
	})
}

func TestListPeopleToNotify(t *testing.T) {
	app := page.AppData{SessionID: "abc"}
	headingLevel := 3
	peopleToNotify := actor.PeopleToNotify{{}}

	want := map[string]interface{}{
		"App":            app,
		"HeadingLevel":   headingLevel,
		"PeopleToNotify": peopleToNotify,
		"CanChange":      true,
	}

	got := listPeopleToNotify(app, peopleToNotify, headingLevel, true)

	assert.Equal(t, want, got)
}

func TestCard(t *testing.T) {
	app := page.AppData{SessionID: "abc"}

	want := map[string]interface{}{
		"App":  app,
		"Item": "hey",
	}

	got := card(app, "hey")

	assert.Equal(t, want, got)
}

func TestPrintStruct(t *testing.T) {
	type str struct {
		Prop1 string
		Prop2 string
	}

	s := str{Prop1: "123"}

	assert.Equal(t, "<p>Prop1: 123</p><p>Prop2: </p>", printStruct(s))
	assert.Equal(t, "<p>Prop1: 123</p><p>Prop2: </p>", printStruct(&s))
}

func TestPossessive(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json")
	app := page.AppData{
		Localizer: bundle.For(localize.En),
	}

	assert.Equal(t, "John’s", possessive(app, "John"))
}

func TestConcatAnd(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json")
	app := page.AppData{
		Localizer: bundle.For(localize.En),
	}

	assert.Equal(t, "", concatAnd(app, []string{}))
	assert.Equal(t, "a", concatAnd(app, []string{"a"}))
	assert.Equal(t, "a and b", concatAnd(app, []string{"a", "b"}))
	assert.Equal(t, "a, b and c", concatAnd(app, []string{"a", "b", "c"}))
}

func TestConcatOr(t *testing.T) {
	bundle, _ := localize.NewBundle("testdata/en.json")
	app := page.AppData{
		Localizer: bundle.For(localize.En),
	}

	assert.Equal(t, "", concatOr(app, []string{}))
	assert.Equal(t, "a", concatOr(app, []string{"a"}))
	assert.Equal(t, "a or b", concatOr(app, []string{"a", "b"}))
	assert.Equal(t, "a, b or c", concatOr(app, []string{"a", "b", "c"}))
}

func TestConcatComma(t *testing.T) {
	assert.Equal(t, "", concatComma([]string{}))
	assert.Equal(t, "a", concatComma([]string{"a"}))
	assert.Equal(t, "a, b", concatComma([]string{"a", "b"}))
	assert.Equal(t, "a, b, c", concatComma([]string{"a", "b", "c"}))
}

func TestContent(t *testing.T) {
	app := page.AppData{SessionID: "abc"}
	componentContent := "content"

	v := content(app, componentContent)

	assert.Equal(t, app, v["App"])
	assert.Equal(t, componentContent, v["Content"])
}

func TestNotificationBanner(t *testing.T) {
	app := page.AppData{SessionID: "abc"}

	assert.Equal(t, notificationBannerData{
		App:     app,
		Title:   "title",
		Content: "content",
		Success: true,
		Heading: true,
	}, notificationBanner(app, "title", "content", "heading", "success"))

	assert.Equal(t, notificationBannerData{
		App:     app,
		Title:   "title",
		Content: "content",
	}, notificationBanner(app, "title", "content"))
}

func TestLpaDecisions(t *testing.T) {
	app := page.AppData{SessionID: "abc"}

	assert.Equal(t, lpaDecisionsData{
		App:       app,
		Lpa:       &lpastore.Lpa{},
		CanChange: true,
	}, lpaDecisions(app, &lpastore.Lpa{}, true))
}

func TestLpaDecisionsWithDonorProvidedDetails(t *testing.T) {
	app := page.AppData{SessionID: "abc"}

	assert.Equal(t, lpaDecisionsData{
		App:       app,
		Lpa:       &lpastore.Lpa{},
		CanChange: true,
	}, lpaDecisions(app, &actor.DonorProvidedDetails{}, true))
}

func TestSummaryRow(t *testing.T) {
	app := page.AppData{SessionID: "abc"}
	label := "a-label"
	value := "aValue"
	changeLink := "a-link.com"
	fullName := "Full Name"

	assert.Equal(t, map[string]any{
		"App":             app,
		"Label":           label,
		"Value":           value,
		"ChangeLink":      changeLink,
		"FullName":        fullName,
		"CanChange":       true,
		"SummarisingSelf": true,
	}, summaryRow(app, label, value, changeLink, fullName, true, true))
}
