// Package templatefn provides functions for use in templates.
package templatefn

import (
	"fmt"
	"html/template"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
)

// Globals contains values that are used in templates and do not change as the
// app runs.
type Globals struct {
	DevMode     bool
	Tag         string
	Region      string
	OneloginURL string
	StaticHash  string
	RumConfig   RumConfig
	ActorTypes  actor.Types
	Paths       appPaths
}

type RumConfig struct {
	GuestRoleArn      string
	Endpoint          string
	ApplicationRegion string
	IdentityPoolID    string
	ApplicationID     string
}

func All(globals *Globals) map[string]any {
	globals.Paths = paths

	return map[string]any{
		"global":             func() *Globals { return globals },
		"isEnglish":          isEnglish,
		"isWelsh":            isWelsh,
		"input":              input,
		"button":             button,
		"items":              items,
		"item":               item,
		"fieldID":            fieldID,
		"errorMessage":       errorMessage,
		"details":            details,
		"inc":                inc,
		"link":               link,
		"fromLink":           fromLink,
		"fromLinkActor":      fromLinkActor,
		"stringContains":     strings.Contains,
		"tr":                 tr,
		"trFormat":           trFormat,
		"trFormatHtml":       trFormatHtml,
		"trHtml":             trHtml,
		"trCount":            trCount,
		"trFormatCount":      trFormatCount,
		"now":                now,
		"addDays":            addDays,
		"formatDate":         formatDate,
		"formatTime":         formatTime,
		"formatDateTime":     formatDateTime,
		"formatPhone":        formatPhone,
		"lowerFirst":         localize.LowerFirst,
		"listAttorneys":      listAttorneys,
		"listPeopleToNotify": listPeopleToNotify,
		"possessive":         possessive,
		"card":               card,
		"printStruct":        printStruct,
		"concatAnd":          concatAnd,
		"concatOr":           concatOr,
		"concatComma":        concatComma,
		"donorCanGoTo":       donor.CanGoTo,
		"content":            content,
		"notificationBanner": notificationBanner,
		"checkboxEq":         checkboxEq,
		"lpaDecisions":       lpaDecisions,
		"summaryRow":         summaryRow,
		"voucherCanGoTo":     voucher.CanGoTo,
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

func button(app appcontext.Data, label string, attrs ...any) map[string]any {
	field := map[string]any{
		"app":   app,
		"label": label,
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

func link(app appcontext.Data, path string) string {
	return app.Lang.URL(path)
}

type lpaIDPath interface{ Format(string) string }

func fromLink(app appcontext.Data, path lpaIDPath, field string) string {
	return app.Lang.URL(path.Format(app.LpaID)) + "?from=" + app.Page + field
}

func fromLinkActor(app appcontext.Data, path lpaIDPath, uid actoruid.UID, field string) string {
	return app.Lang.URL(path.Format(app.LpaID)) + "?from=" + app.Page + "&id=" + uid.String() + field
}

// checkboxEq allows matching in the checkboxes.gohtml template for a value that
// is a list of strings, or a single string (where we are emulating a switch)
func checkboxEq(needle string, in any) bool {
	if in == nil {
		return false
	}

	if str, ok := in.(string); ok {
		return needle == str
	}

	if slist, ok := in.([]string); ok {
		return slices.Contains(slist, needle)
	}

	return false
}

func tr(app appcontext.Data, messageID string) string {
	if messageID == "" {
		return ""
	}

	return app.Localizer.T(messageID)
}

func trFormat(app appcontext.Data, messageID string, args ...interface{}) string {
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

func trFormatHtml(app appcontext.Data, messageID string, args ...interface{}) template.HTML {
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

func trHtml(app appcontext.Data, messageID string) template.HTML {
	if messageID == "" {
		return ""
	}

	return template.HTML(app.Localizer.T(messageID))
}

func trCount(app appcontext.Data, messageID string, count int) string {
	if messageID == "" {
		return ""
	}

	return app.Localizer.Count(messageID, count)
}

func trFormatCount(app appcontext.Data, messageID string, count int, args ...interface{}) string {
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

func formatDate(app appcontext.Data, t date.TimeOrDate) string {
	return app.Localizer.FormatDate(t)
}

func formatTime(app appcontext.Data, t time.Time) string {
	return app.Localizer.FormatTime(t)
}

func formatDateTime(app appcontext.Data, t time.Time) string {
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

type attorneySummaryData struct {
	App              appcontext.Data
	CanChange        bool
	TrustCorporation lpadata.TrustCorporation
	Attorneys        []lpadata.Attorney
	Link             attorneySummaryDataLinks
	HeadingLevel     int
}

type attorneySummaryDataLinks struct {
	TrustCorporation, TrustCorporationAddress, RemoveTrustCorporation string
	Attorney, AttorneyAddress, RemoveAttorney                         string
}

func listAttorneys(app appcontext.Data, attorneys any, attorneyType string, headingLevel int, canChange bool) attorneySummaryData {
	data := attorneySummaryData{
		App:          app,
		CanChange:    canChange,
		HeadingLevel: headingLevel,
	}

	switch v := attorneys.(type) {
	case lpadata.Attorneys:
		data.Attorneys = v.Attorneys
		data.TrustCorporation = v.TrustCorporation
	case donordata.Attorneys:
		for _, a := range v.Attorneys {
			data.Attorneys = append(data.Attorneys, lpadata.Attorney{
				UID:         a.UID,
				FirstNames:  a.FirstNames,
				LastName:    a.LastName,
				DateOfBirth: a.DateOfBirth,
				Email:       a.Email,
				Address:     a.Address,
			})
		}

		if t := v.TrustCorporation; t.Name != "" {
			data.TrustCorporation = lpadata.TrustCorporation{
				UID:           t.UID,
				Name:          t.Name,
				CompanyNumber: t.CompanyNumber,
				Email:         t.Email,
				Address:       t.Address,
			}
		}
	default:
		panic("unsupported type of attorneys for listAttorneys")
	}

	if attorneyType == "replacement" {
		data.Link.Attorney = fmt.Sprintf("%s?from=%s", donor.PathChooseReplacementAttorneys.Format(app.LpaID), app.Page)
		data.Link.AttorneyAddress = fmt.Sprintf("%s?from=%s", donor.PathChooseReplacementAttorneysAddress.Format(app.LpaID), app.Page)
		data.Link.RemoveAttorney = fmt.Sprintf("%s?from=%s", donor.PathRemoveReplacementAttorney.Format(app.LpaID), app.Page)
		data.Link.TrustCorporation = fmt.Sprintf("%s?from=%s", donor.PathEnterReplacementTrustCorporation.Format(app.LpaID), app.Page)
		data.Link.TrustCorporationAddress = fmt.Sprintf("%s?from=%s", donor.PathEnterReplacementTrustCorporationAddress.Format(app.LpaID), app.Page)
		data.Link.RemoveTrustCorporation = fmt.Sprintf("%s?from=%s", donor.PathRemoveReplacementTrustCorporation.Format(app.LpaID), app.Page)
	} else {
		data.Link.Attorney = fmt.Sprintf("%s?from=%s", donor.PathChooseAttorneys.Format(app.LpaID), app.Page)
		data.Link.AttorneyAddress = fmt.Sprintf("%s?from=%s", donor.PathChooseAttorneysAddress.Format(app.LpaID), app.Page)
		data.Link.RemoveAttorney = fmt.Sprintf("%s?from=%s", donor.PathRemoveAttorney.Format(app.LpaID), app.Page)
		data.Link.TrustCorporation = fmt.Sprintf("%s?from=%s", donor.PathEnterTrustCorporation.Format(app.LpaID), app.Page)
		data.Link.TrustCorporationAddress = fmt.Sprintf("%s?from=%s", donor.PathEnterTrustCorporationAddress.Format(app.LpaID), app.Page)
		data.Link.RemoveTrustCorporation = fmt.Sprintf("%s?from=%s", donor.PathRemoveTrustCorporation.Format(app.LpaID), app.Page)
	}

	return data
}

func listPeopleToNotify(app appcontext.Data, peopleToNotify any, headingLevel int, canChange bool) map[string]interface{} {
	return map[string]any{
		"App":            app,
		"HeadingLevel":   headingLevel,
		"PeopleToNotify": peopleToNotify,
		"CanChange":      canChange,
	}
}

func card(app appcontext.Data, item any) map[string]any {
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

func possessive(app appcontext.Data, s string) string {
	return app.Localizer.Possessive(s)
}

func concatAnd(app appcontext.Data, list []string) string {
	return app.Localizer.Concat(list, "and")
}

func concatOr(app appcontext.Data, list []string) string {
	return app.Localizer.Concat(list, "or")
}

func concatComma(list []string) string {
	return strings.Join(list, ", ")
}

func content(app appcontext.Data, content string) map[string]interface{} {
	return map[string]interface{}{
		"App":     app,
		"Content": content,
	}
}

type notificationBannerData struct {
	App     appcontext.Data
	Title   string
	Content template.HTML
	Heading bool
	Success bool
}

func notificationBanner(app appcontext.Data, title string, content template.HTML, options ...string) notificationBannerData {
	return notificationBannerData{
		App:     app,
		Title:   title,
		Content: content,
		Heading: slices.Contains(options, "heading"),
		Success: slices.Contains(options, "success"),
	}
}

type lpaDecisionsData struct {
	App       appcontext.Data
	Lpa       *lpadata.Lpa
	CanChange bool
}

func lpaDecisions(app appcontext.Data, lpa any, canChange bool) lpaDecisionsData {
	data := lpaDecisionsData{
		App:       app,
		CanChange: canChange,
	}

	switch v := lpa.(type) {
	case *lpadata.Lpa:
		data.Lpa = v
	case *donordata.Provided:
		data.Lpa = lpastore.FromDonorProvidedDetails(v)
	}

	return data
}

func summaryRow(app appcontext.Data, label string, value any, changeLink, fullName string, canChange, summarisingSelf bool) map[string]any {
	return map[string]any{
		"App":             app,
		"Label":           label,
		"Value":           value,
		"ChangeLink":      changeLink,
		"FullName":        fullName,
		"CanChange":       canChange,
		"SummarisingSelf": summarisingSelf,
	}
}
