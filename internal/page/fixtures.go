package page

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
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
			var values url.Values

			switch data.Form.Journey {
			case "attorney":
				values = url.Values{
					"useTestShareCode":            {"1"},
					"sendAttorneyShare":           {"1"},
					"lpa.complete":                {"1"},
					"lpa.attorneys":               {"2"},
					"lpa.attorneysAct":            {actor.JointlyAndSeverally.String()},
					"lpa.replacementAttorneys":    {"2"},
					"lpa.replacementAttorneysAct": {actor.Jointly.String()},
					"lpa.type":                    {data.Form.Type},
					"lpa.restrictions":            {"1"},
					"redirect":                    {Paths.Attorney.Start.Format()},
					"lpa.progress":                {r.FormValue("lpa.progress")},
				}

				switch data.Form.SendTo {
				case "replacement-attorney":
					if data.Form.Email != "" {
						values.Add("lpa.replacementAttorneyEmail", data.Form.Email)
					}
				case "trust-corporation":
					values.Add("lpa.trustCorporation", "complete")
					if data.Form.Email != "" {
						values.Add("lpa.trustCorporationEmail", data.Form.Email)
					}
				case "replacement-trust-corporation":
					values.Add("lpa.replacementTrustCorporation", "complete")
					if data.Form.Email != "" {
						values.Add("lpa.replacementTrustCorporationEmail", data.Form.Email)
					}
				default:
					if data.Form.Email != "" {
						values.Add("lpa.attorneyEmail", data.Form.Email)
					}
				}

				if data.Form.Signed != "" {
					values.Add("lpa.signedByDonor", "1")
					values.Add("asCertificateProvider", "certified")
				}

			case "certificate-provider":
				values = url.Values{
					"useTestShareCode":  {"1"},
					data.Form.DonorPaid: {"1"},
				}

				if data.Form.Email != "" {
					values.Add("lpa.certificateProviderEmail", data.Form.Email)
				}

				if data.Form.DonorPaid != "" {
					values.Add("startCpFlowDonorHasPaid", "1")
				} else {
					values.Add("startCpFlowDonorHasNotPaid", "1")
				}

				if data.Form.Signed != "" {
					values.Add("lpa.signedByDonor", "1")
				}

			case "donor":
				values = url.Values{
					"lpa.type":                    {data.Form.Type},
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
					values.Add("lpa.attorneys", data.Form.AttorneyCount)
				}

				if data.Form.ReplacementAttorneys != "" {
					values.Add("lpa.replacementAttorneys", data.Form.ReplacementAttorneyCount)
				}

				if data.Form.PeopleToNotify != "" {
					values.Add("lpa.peopleToNotify", data.Form.PersonToNotifyCount)
				}
			case "everything":
				values = url.Values{
					"fresh":                 {"1"},
					"redirect":              {Paths.Dashboard.Format()},
					"attorneyProvided":      {r.FormValue("attorneyProvided")},
					"asCertificateProvider": {r.FormValue("asCertificateProvider")},
					"lpa.progress":          {r.FormValue("lpa.progress")},
				}

				if r.FormValue("lpa.progress") == "" {
					values.Set("lpa.yourDetails", "1")
				} else {
					values.Set("lpa.complete", "1")
				}
			}

			http.Redirect(w, r, fmt.Sprintf("%s?%s", Paths.TestingStart, values.Encode()), http.StatusFound)
			return nil
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
	SendTo                   string
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
		SendTo:                   PostFormString(r, "send-to"),
		Signed:                   PostFormString(r, "signed"),
		Type:                     PostFormString(r, "type"),
	}
}
