package page

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type fixtureData struct {
	App    AppData
	Errors validation.List
	Form   *fixturesForm
}

func Fixtures(tmpl template.Template) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &fixtureData{
			App:  appData,
			Form: &fixturesForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readFixtures(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				var values url.Values

				switch data.Form.Journey {
				case "attorney":
					values = url.Values{
						"useTestShareCode":           {"1"},
						"sendAttorneyShare":          {"1"},
						"completeLpa":                {"1"},
						"withAttorneys":              {"1"},
						"howAttorneysAct":            {"jointly-and-severally"},
						"withReplacementAttorneys":   {"1"},
						"howReplacementAttorneysAct": {"jointly"},
						"withType":                   {data.Form.Type},
						"withRestrictions":           {"1"},
						"redirect":                   {Paths.Attorney.Start},
					}
					if data.Form.Email != "" {
						values.Add("withEmail", data.Form.Email)
					}
					if data.Form.ForReplacementAttorney != "" {
						values.Add("forReplacementAttorney", "1")
					}
					if data.Form.Signed != "" {
						values.Add("signedByDonor", "1")
						values.Add("provideCertificate", "1")
					}

				case "certificate-provider":
					values = url.Values{
						"useTestShareCode":  {"1"},
						data.Form.DonorPaid: {"1"},
					}

					if data.Form.Email != "" {
						values.Add("withEmail", data.Form.Email)
					}

					if data.Form.DonorPaid != "" {
						values.Add("startCpFlowDonorHasPaid", "1")
					} else {
						values.Add("startCpFlowDonorHasNotPaid", "1")
					}

				case "donor":
					values = url.Values{
						data.Form.DonorDetails:        {"1"},
						data.Form.WhenCanLpaBeUsed:    {"1"},
						data.Form.Restrictions:        {"1"},
						data.Form.CertificateProvider: {"1"},
						data.Form.CheckAndSend:        {"1"},
						data.Form.Pay:                 {"1"},
						data.Form.IdAndSign:           {"1"},
						data.Form.CompleteAll:         {"1"},
					}

					if data.Form.Attorneys != "" {
						values.Add("withAttorneys", data.Form.AttorneyCount)
					}

					if data.Form.ReplacementAttorneys != "" {
						values.Add("withReplacementAttorneys", data.Form.ReplacementAttorneyCount)
					}

					if data.Form.PeopleToNotify != "" {
						values.Add("withPeopleToNotify", data.Form.PersonToNotifyCount)
					}
				case "everything":
					values = url.Values{"fresh": {"1"}, "completeLpa": {"1"}, "redirect": {Paths.Dashboard}}

					if r.FormValue("as-attorney") != "" {
						values.Add("asAttorney", "1")
					}

					if r.FormValue("as-certificate-provider") != "" {
						values.Add("asCertificateProvider", "1")
					}
				}

				http.Redirect(w, r, fmt.Sprintf("%s?%s", Paths.TestingStart, values.Encode()), http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type fixturesForm struct {
	Journey                  string
	DonorDetails             string
	Attorneys                string
	AttorneyCount            string
	ReplacementAttorneys     string
	ReplacementAttorneyCount string
	WhenCanLpaBeUsed         string
	Restrictions             string
	CertificateProvider      string
	PeopleToNotify           string
	PersonToNotifyCount      string
	CheckAndSend             string
	Pay                      string
	IdAndSign                string
	CompleteAll              string
	Email                    string
	DonorPaid                string
	ForReplacementAttorney   string
	Signed                   string
	Type                     string
}

func readFixtures(r *http.Request) *fixturesForm {
	return &fixturesForm{
		Journey:                  PostFormString(r, "journey"),
		DonorDetails:             PostFormString(r, "donor-details"),
		Attorneys:                PostFormString(r, "attorneys"),
		AttorneyCount:            PostFormString(r, "attorney-count"),
		ReplacementAttorneys:     PostFormString(r, "replacement-attorneys"),
		ReplacementAttorneyCount: PostFormString(r, "replacement-attorney-count"),
		WhenCanLpaBeUsed:         PostFormString(r, "when-can-lpa-be-used"),
		Restrictions:             PostFormString(r, "restrictions"),
		CertificateProvider:      PostFormString(r, "certificate-provider"),
		PeopleToNotify:           PostFormString(r, "people-to-notify"),
		PersonToNotifyCount:      PostFormString(r, "person-to-notify-count"),
		CheckAndSend:             PostFormString(r, "check-and-send-to-cp"),
		Pay:                      PostFormString(r, "pay-for-lpa"),
		IdAndSign:                PostFormString(r, "confirm-id-and-sign"),
		CompleteAll:              PostFormString(r, "complete-all-sections"),
		Email:                    PostFormString(r, "email"),
		DonorPaid:                PostFormString(r, "donor-paid"),
		ForReplacementAttorney:   PostFormString(r, "for-replacement-attorney"),
		Signed:                   PostFormString(r, "signed"),
		Type:                     PostFormString(r, "type"),
	}
}

func (f *fixturesForm) Validate() validation.List {
	var errors validation.List

	if f.Journey == "certificate-provider" && f.Email != "" && f.DonorPaid == "" {
		errors.String("cp-flow-has-donor-paid", "how to start the CP flow", f.DonorPaid,
			validation.Select("startCpFlowWithId", "startCpFlowWithoutId"))
	}

	return errors
}
