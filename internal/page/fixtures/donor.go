package fixtures

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func Donor(tmpl template.Template) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := &fixturesData{
			App: appData,
		}

		if r.Method == http.MethodPost {
			form := readFixtures(r)
			var values url.Values

			switch form.Journey {
			case "donor":
				values = url.Values{
					"lpa.type":               {form.Type},
					form.DonorDetails:        {"1"},
					form.WhenCanLpaBeUsed:    {"1"},
					form.Restrictions:        {"1"},
					form.CertificateProvider: {"1"},
					form.CheckAndSend:        {"1"},
					form.Pay:                 {"1"},
					form.IdAndSign:           {"1"},
					form.CompleteAll:         {"1"},
				}

				if form.Attorneys != "" {
					values.Add("lpa.attorneys", form.AttorneyCount)
				}

				if form.ReplacementAttorneys != "" {
					values.Add("lpa.replacementAttorneys", form.ReplacementAttorneyCount)
				}

				if form.PeopleToNotify != "" {
					values.Add("lpa.peopleToNotify", form.PersonToNotifyCount)
				}
			}

			http.Redirect(w, r, fmt.Sprintf("%s?%s", page.Paths.TestingStart, values.Encode()), http.StatusFound)
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
	SendTo                   string
	Signed                   string
	Type                     string
}

func readFixtures(r *http.Request) *fixturesForm {
	return &fixturesForm{
		Journey:                  r.FormValue("journey"),
		DonorDetails:             r.FormValue("donor-details"),
		Attorneys:                r.FormValue("attorneys"),
		AttorneyCount:            r.FormValue("attorney-count"),
		ReplacementAttorneys:     r.FormValue("replacement-attorneys"),
		ReplacementAttorneyCount: r.FormValue("replacement-attorney-count"),
		WhenCanLpaBeUsed:         r.FormValue("when-can-lpa-be-used"),
		Restrictions:             r.FormValue("restrictions"),
		CertificateProvider:      r.FormValue("certificate-provider"),
		PeopleToNotify:           r.FormValue("people-to-notify"),
		PersonToNotifyCount:      r.FormValue("person-to-notify-count"),
		CheckAndSend:             r.FormValue("check-and-send-to-cp"),
		Pay:                      r.FormValue("pay-for-lpa"),
		IdAndSign:                r.FormValue("confirm-id-and-sign"),
		CompleteAll:              r.FormValue("complete-all-sections"),
		Email:                    r.FormValue("email"),
		SendTo:                   r.FormValue("send-to"),
		Signed:                   r.FormValue("signed"),
		Type:                     r.FormValue("type"),
	}
}
