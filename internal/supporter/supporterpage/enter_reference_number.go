package supporterpage

import (
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReferenceNumber struct {
	App    appcontext.Data
	Errors validation.List
	Form   *referenceNumberForm
}

func EnterReferenceNumber(logger Logger, tmpl template.Template, memberStore MemberStore, sessionStore SessionStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := &enterReferenceNumber{
			App: appData,
			Form: &referenceNumberForm{
				Label: "referenceNumber",
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readReferenceNumberForm(r, "referenceNumber")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				invite, err := memberStore.InvitedMember(r.Context())
				if err != nil {
					return err
				}

				if invite.ReferenceNumber != data.Form.ReferenceNumber {
					data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"})
					return tmpl(w, data)
				}

				if invite.HasExpired() {
					return page.PathSupporterInviteExpired.Redirect(w, r, appData)
				}

				if err := memberStore.CreateFromInvite(r.Context(), invite); err != nil {
					return err
				}

				loginSession, err := sessionStore.Login(r)
				if err != nil {
					return page.PathSupporterStart.Redirect(w, r, appData)
				}

				loginSession.OrganisationID = invite.OrganisationID
				loginSession.OrganisationName = invite.OrganisationName

				logger.InfoContext(r.Context(), "member invite redeemed", slog.String("organisation_id", loginSession.OrganisationID))

				if err := sessionStore.SetLogin(r, w, loginSession); err != nil {
					return err
				}

				return supporter.PathDashboard.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}

type referenceNumberForm struct {
	ReferenceNumber    string
	ReferenceNumberRaw string
	Label              string
}

func readReferenceNumberForm(r *http.Request, label string) *referenceNumberForm {
	return &referenceNumberForm{
		ReferenceNumber:    page.PostFormReferenceNumber(r, "reference-number"),
		ReferenceNumberRaw: page.PostFormString(r, "reference-number"),
		Label:              label,
	}
}

func (f *referenceNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("reference-number", "twelveCharactersReferenceNumber", f.ReferenceNumber,
		validation.Empty())

	errors.String("reference-number", "theReferenceNumberYouEnter", f.ReferenceNumber,
		validation.StringLength(12))

	return errors
}
