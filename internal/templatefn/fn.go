package templatefn

import (
	"fmt"
	"html/template"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	humanize "github.com/dustin/go-humanize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func All(tag, region string) map[string]any {
	return map[string]any{
		"buildTag":           func() string { return tag },
		"awsRegion":          func() string { return region },
		"isEnglish":          isEnglish,
		"isWelsh":            isWelsh,
		"input":              input,
		"items":              items,
		"item":               item,
		"fieldID":            fieldID,
		"errorMessage":       errorMessage,
		"details":            details,
		"inc":                inc,
		"link":               link,
		"contains":           contains,
		"tr":                 tr,
		"trFormat":           trFormat,
		"trFormatHtml":       trFormatHtml,
		"trHtml":             trHtml,
		"trCount":            trCount,
		"trFormatCount":      trFormatCount,
		"now":                now,
		"addDays":            addDays,
		"formatDate":         formatDate,
		"formatDateTime":     formatDateTime,
		"formatPhone":        formatPhone,
		"lowerFirst":         lowerFirst,
		"listAttorneys":      listAttorneys,
		"warning":            warning,
		"listPeopleToNotify": listPeopleToNotify,
		"possessive":         possessive,
		"card":               card,
		"printStruct":        printStruct,
		"concatAnd":          concatAnd,
		"concatOr":           concatOr,
		"concatComma":        concatComma,
		"penceToPounds":      penceToPounds,
		"canGoTo":            page.CanGoTo,
	}
}

func isEnglish(lang localize.Lang) bool {
	return lang == localize.En
}

func isWelsh(lang localize.Lang) bool {
	return lang == localize.Cy
}

func input(top interface{}, name, label string, value interface{}, attrs ...interface{}) map[string]interface{} {
	field := map[string]interface{}{
		"top":   top,
		"name":  name,
		"label": label,
		"value": value,
	}

	if len(attrs)%2 != 0 {
		panic("must have even number of attrs")
	}

	for i := 0; i < len(attrs); i += 2 {
		field[attrs[i].(string)] = attrs[i+1]
	}

	return field
}

func items(top interface{}, name string, value interface{}, items ...interface{}) map[string]interface{} {
	return map[string]interface{}{
		"top":   top,
		"name":  name,
		"value": value,
		"items": items,
	}
}

func item(value, label string, attrs ...interface{}) map[string]interface{} {
	item := map[string]interface{}{
		"value": value,
		"label": label,
	}

	if len(attrs)%2 != 0 {
		panic("must have even number of attrs")
	}

	for i := 0; i < len(attrs); i += 2 {
		item[attrs[i].(string)] = attrs[i+1]
	}

	return item
}

func fieldID(name string, i int) string {
	if i == 0 {
		return name
	}

	return fmt.Sprintf("%s-%d", name, i+1)
}

func errorMessage(top interface{}, name string) map[string]interface{} {
	return map[string]interface{}{
		"top":  top,
		"name": name,
	}
}

func details(top interface{}, name, detail string, open bool) map[string]interface{} {
	return map[string]interface{}{
		"top":    top,
		"name":   name,
		"detail": detail,
		"open":   open,
	}
}

func inc(i int) int {
	return i + 1
}

func link(app page.AppData, path string) string {
	return app.Lang.URL(path)
}

func contains(needle string, list any) bool {
	if list == nil {
		return false
	}

	if slist, ok := list.([]string); ok {
		return slices.Contains(slist, needle)
	}

	return false
}

func tr(app page.AppData, messageID string) string {
	if messageID == "" {
		return ""
	}

	return app.Localizer.T(messageID)
}

func trFormat(app page.AppData, messageID string, args ...interface{}) string {
	if messageID == "" {
		return ""
	}

	if len(args)%2 != 0 {
		panic("must have even number of args")
	}

	data := map[string]interface{}{}
	for i := 0; i < len(args); i += 2 {
		data[args[i].(string)] = args[i+1]
	}

	return app.Localizer.Format(messageID, data)
}

func trFormatHtml(app page.AppData, messageID string, args ...interface{}) template.HTML {
	if messageID == "" {
		return ""
	}

	if len(args)%2 != 0 {
		panic("must have even number of args")
	}

	data := map[string]interface{}{}
	for i := 0; i < len(args); i += 2 {
		data[args[i].(string)] = args[i+1]
	}

	return template.HTML(app.Localizer.Format(messageID, data))
}

func trHtml(app page.AppData, messageID string) template.HTML {
	if messageID == "" {
		return ""
	}

	return template.HTML(app.Localizer.T(messageID))
}

func trCount(app page.AppData, messageID string, count int) string {
	if messageID == "" {
		return ""
	}

	return app.Localizer.Count(messageID, count)
}

func trFormatCount(app page.AppData, messageID string, count int, args ...interface{}) string {
	if messageID == "" {
		return ""
	}

	if len(args)%2 != 0 {
		panic("must have even number of args")
	}

	data := map[string]interface{}{}
	for i := 0; i < len(args); i += 2 {
		data[args[i].(string)] = args[i+1]
	}

	return app.Localizer.FormatCount(messageID, count, data)
}

func now() time.Time {
	return time.Now()
}

func addDays(days int, t time.Time) time.Time {
	return t.AddDate(0, 0, days)
}

func formatDate(app page.AppData, t date.TimeOrDate) string {
	return app.Localizer.FormatDate(t)
}

func formatDateTime(app page.AppData, t time.Time) string {
	return app.Localizer.FormatDateTime(t)
}

func formatPhone(s string) string {
	stripped := strings.ReplaceAll(s, " ", "")

	// 07005 060 702
	if len(stripped) == 11 && stripped[0] == '0' {
		return stripped[:5] + " " + stripped[5:8] + " " + stripped[8:]
	}

	// +44 7005 060 702
	if len(stripped) == 13 && stripped[:3] == "+44" {
		return stripped[:3] + " " + stripped[3:7] + " " + stripped[7:10] + " " + stripped[10:]
	}

	return s
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

type attorneySummaryData struct {
	App              page.AppData
	CanChange        bool
	TrustCorporation actor.TrustCorporation
	Attorneys        []actor.Attorney
	Link             struct {
		TrustCorporation, TrustCorporationAddress, RemoveTrustCorporation string
		Attorney, AttorneyAddress, RemoveAttorney                         string
	}
	HeadingLevel int
}

func listAttorneys(attorneys actor.Attorneys, app page.AppData, attorneyType string, headingLevel int, donor *actor.DonorProvidedDetails) attorneySummaryData {
	data := attorneySummaryData{
		App:              app,
		CanChange:        !donor.Tasks.ConfirmYourIdentityAndSign.Completed() && app.IsDonor(),
		TrustCorporation: attorneys.TrustCorporation,
		Attorneys:        attorneys.Attorneys,
		HeadingLevel:     headingLevel,
	}

	if attorneyType == "replacement" {
		data.Link.Attorney = fmt.Sprintf("%s?from=%s", app.Paths.ChooseReplacementAttorneys.Format(donor.LpaID), app.Page)
		data.Link.AttorneyAddress = fmt.Sprintf("%s?from=%s", app.Paths.ChooseReplacementAttorneysAddress.Format(donor.LpaID), app.Page)
		data.Link.RemoveAttorney = fmt.Sprintf("%s?from=%s", app.Paths.RemoveReplacementAttorney.Format(donor.LpaID), app.Page)
		data.Link.TrustCorporation = fmt.Sprintf("%s?from=%s", app.Paths.EnterReplacementTrustCorporation.Format(donor.LpaID), app.Page)
		data.Link.TrustCorporationAddress = fmt.Sprintf("%s?from=%s", app.Paths.EnterReplacementTrustCorporationAddress.Format(donor.LpaID), app.Page)
		data.Link.RemoveTrustCorporation = fmt.Sprintf("%s?from=%s", app.Paths.RemoveReplacementTrustCorporation.Format(donor.LpaID), app.Page)
	} else {
		data.Link.Attorney = fmt.Sprintf("%s?from=%s", app.Paths.ChooseAttorneys.Format(donor.LpaID), app.Page)
		data.Link.AttorneyAddress = fmt.Sprintf("%s?from=%s", app.Paths.ChooseAttorneysAddress.Format(donor.LpaID), app.Page)
		data.Link.RemoveAttorney = fmt.Sprintf("%s?from=%s", app.Paths.RemoveAttorney.Format(donor.LpaID), app.Page)
		data.Link.TrustCorporation = fmt.Sprintf("%s?from=%s", app.Paths.EnterTrustCorporation.Format(donor.LpaID), app.Page)
		data.Link.TrustCorporationAddress = fmt.Sprintf("%s?from=%s", app.Paths.EnterTrustCorporationAddress.Format(donor.LpaID), app.Page)
		data.Link.RemoveTrustCorporation = fmt.Sprintf("%s?from=%s", app.Paths.RemoveTrustCorporation.Format(donor.LpaID), app.Page)
	}

	return data
}

func listPeopleToNotify(app page.AppData, headingLevel int, donor *actor.DonorProvidedDetails) map[string]interface{} {
	return map[string]interface{}{
		"App":          app,
		"HeadingLevel": headingLevel,
		"Donor":        donor,
	}
}

func warning(app page.AppData, content string) map[string]interface{} {
	return map[string]interface{}{
		"app":     app,
		"content": content,
	}
}

func card(app page.AppData, item any) map[string]any {
	return map[string]interface{}{
		"App":  app,
		"Item": item,
	}
}

// printStruct is a quick way to print out an easy-to-read text representation of a struct in a template:
// {{ trHtml .App (printStruct .Lpa) }}
func printStruct(s interface{}) string {
	v := reflect.ValueOf(s)
	typeOfS := v.Type()
	var output string

	if typeOfS.Kind() == reflect.Ptr {
		for i := 0; i < v.Elem().NumField(); i++ {
			output = output + fmt.Sprintf("<p>%s: %v</p>", typeOfS.Elem().Field(i).Name, v.Elem().Field(i).Interface())
		}
	} else {
		for i := 0; i < v.NumField(); i++ {
			output = output + fmt.Sprintf("<p>%s: %v</p>", typeOfS.Field(i).Name, v.Field(i).Interface())
		}
	}

	return output
}

func possessive(app page.AppData, s string) string {
	return app.Localizer.Possessive(s)
}

func concatAnd(app page.AppData, list []string) string {
	return app.Localizer.Concat(list, "and")
}

func concatOr(app page.AppData, list []string) string {
	return app.Localizer.Concat(list, "or")
}

func concatComma(list []string) string {
	return strings.Join(list, ", ")
}

func penceToPounds(pence int) string {
	return humanize.CommafWithDigits(float64(pence)/100, 2)
}
