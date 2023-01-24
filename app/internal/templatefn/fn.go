package templatefn

import (
	"fmt"
	"html/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"golang.org/x/exp/slices"
)

var All = map[string]interface{}{
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
	"lowerFirst":         lowerFirst,
	"listAttorneys":      listAttorneys,
	"warning":            warning,
	"listPeopleToNotify": listPeopleToNotify,
	"progressBar":        progressBar,
}

func isEnglish(lang page.Lang) bool {
	return lang == page.En
}

func isWelsh(lang page.Lang) bool {
	return lang == page.Cy
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

func formatDate(t time.Time) string {
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

func listAttorneys(attorneys []page.Attorney, app page.AppData, detailsPath, addressPath, removePath string) map[string]interface{} {
	return map[string]interface{}{
		"Attorneys":   attorneys,
		"App":         app,
		"DetailsPath": detailsPath,
		"AddressPath": addressPath,
		"RemovePath":  removePath,
	}
}

func listPeopleToNotify(peopleToNotify []page.PersonToNotify, app page.AppData, detailsPath, addressPath, removePath string) map[string]interface{} {
	return map[string]interface{}{
		"PeopleToNotify": peopleToNotify,
		"App":            app,
		"DetailsPath":    detailsPath,
		"AddressPath":    addressPath,
		"RemovePath":     removePath,
	}
}

func warning(app page.AppData, content string) map[string]interface{} {
	return map[string]interface{}{
		"app":     app,
		"content": content,
	}
}

func progressBar(app page.AppData, lpa *page.Lpa) map[string]interface{} {
	return map[string]interface{}{
		"App": app,
		"Lpa": lpa,
	}
}
