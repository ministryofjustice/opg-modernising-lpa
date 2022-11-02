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
	"isEnglish":       isEnglish,
	"isWelsh":         isWelsh,
	"input":           input,
	"items":           items,
	"item":            item,
	"fieldID":         fieldID,
	"errorMessage":    errorMessage,
	"details":         details,
	"inc":             inc,
	"link":            link,
	"contains":        contains,
	"tr":              tr,
	"trFormat":        trFormat,
	"trFormatHtml":    trFormatHtml,
	"trHtml":          trHtml,
	"trCount":         trCount,
	"now":             now,
	"addDays":         addDays,
	"formatDate":      formatDate,
	"formatDateTime":  formatDateTime,
	"lowerFirst":      lowerFirst,
	"attorneyDetails": attorneyDetails,
	"join":            join,
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
	if app.Lang == page.Cy {
		return "/cy" + path
	}

	return path
}

func contains(needle string, list interface{}) bool {
	if slist, ok := list.([]string); ok {
		return slices.Contains(slist, needle)
	}

	if slist, ok := list.([]page.IdentityOption); ok {
		for _, item := range slist {
			if item.String() == needle {
				return true
			}
		}
	}

	return false
}

func tr(app page.AppData, messageID string) string {
	return app.Localizer.T(messageID)
}

func trFormat(app page.AppData, messageID string, args ...interface{}) string {
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
	return template.HTML(app.Localizer.T(messageID))
}

func trCount(app page.AppData, messageID string, count int) string {
	return app.Localizer.Count(messageID, count)
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

	return t.Format("15:04:05, 2 January 2006")
}

func lowerFirst(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func attorneyDetails(attorneys []page.Attorney, from string, app page.AppData) map[string]interface{} {
	return map[string]interface{}{
		"Attorneys": attorneys,
		"From":      from,
		"App":       app,
	}
}
