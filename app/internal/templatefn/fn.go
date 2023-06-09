package templatefn

import (
	"fmt"
	"html/template"
	"reflect"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"golang.org/x/exp/slices"
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
		"linkWithID":         linkWithID,
		"linkLang":           linkLang,
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
		"lowerFirst":         lowerFirst,
		"listAttorneys":      listAttorneys,
		"warning":            warning,
		"listPeopleToNotify": listPeopleToNotify,
		"progressBar":        progressBar,
		"peopleNamedOnLpa":   peopleNamedOnLpa,
		"possessive":         possessive,
		"card":               card,
		"printStruct":        printStruct,
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
	return app.BuildUrl(path)
}

func linkWithID(app page.AppData, path, lpaID string) string {
	return page.AppData{Lang: app.Lang, LpaID: lpaID}.BuildUrl(path)
}

func linkLang(app page.AppData, path string) string {
	if app.Lang == localize.Cy {
		return "/" + app.Lang.String() + path
	}

	return path
}

func contains(needle string, list []string) bool {
	return slices.Contains(list, needle)
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

type dateOrTime interface {
	IsZero() bool
	Format(string) string
}

func formatDate(t dateOrTime) string {
	if t.IsZero() {
		return ""
	}

	return t.Format("2 January 2006")
}

func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format("2 January 2006 at 15:04")
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func listAttorneys(attorneys actor.Attorneys, app page.AppData, attorneyType string, withHeaders bool, lpa *page.Lpa) map[string]interface{} {
	props := map[string]interface{}{
		"Attorneys":    attorneys,
		"App":          app,
		"WithHeaders":  withHeaders,
		"Lpa":          lpa,
		"AttorneyType": attorneyType,
	}

	if attorneyType == "replacement" {
		props["DetailsPath"] = fmt.Sprintf("%s?from=%s", app.Paths.ChooseReplacementAttorneys, app.Page)
		props["AddressPath"] = fmt.Sprintf("%s?from=%s", app.Paths.ChooseReplacementAttorneysAddress, app.Page)
		props["RemovePath"] = fmt.Sprintf("%s?from=%s", app.Paths.RemoveReplacementAttorney, app.Page)
	} else {
		props["DetailsPath"] = fmt.Sprintf("%s?from=%s", app.Paths.ChooseAttorneys, app.Page)
		props["AddressPath"] = fmt.Sprintf("%s?from=%s", app.Paths.ChooseAttorneysAddress, app.Page)
		props["RemovePath"] = fmt.Sprintf("%s?from=%s", app.Paths.RemoveAttorney, app.Page)
	}

	return props
}

func listPeopleToNotify(app page.AppData, withHeaders bool, lpa *page.Lpa) map[string]interface{} {
	return map[string]interface{}{
		"App":         app,
		"WithHeaders": withHeaders,
		"Lpa":         lpa,
	}
}

func warning(app page.AppData, content string) map[string]interface{} {
	return map[string]interface{}{
		"app":     app,
		"content": content,
	}
}

func progressBar(app page.AppData, lpa *page.Lpa, certificateProvider *actor.CertificateProviderProvidedDetails) map[string]interface{} {
	return map[string]interface{}{
		"App":                 app,
		"Lpa":                 lpa,
		"CertificateProvider": certificateProvider,
	}
}

func peopleNamedOnLpa(app page.AppData, lpa *page.Lpa, showPeopleHeaders bool) map[string]interface{} {
	return map[string]interface{}{
		"App":               app,
		"Lpa":               lpa,
		"ShowPeopleHeaders": showPeopleHeaders,
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
