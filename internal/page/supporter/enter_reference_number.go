package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReferenceNumber struct {
	App    page.AppData
	Errors validation.List
	Form   *referenceNumberForm
}

func EnterReferenceNumber(tmpl template.Template, memberStore MemberStore, sessionStore sesh.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
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
				invites, err := memberStore.InvitedMembersByEmail(r.Context())
				if err != nil {
					return err
				}

				var invite *actor.MemberInvite
				for _, i := range invites {
					if i.ReferenceNumber == data.Form.ReferenceNumber {
						invite = i
						break
					}
				}

				if invite == nil {
					data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"})
					return tmpl(w, data)
				}

				if invite.HasExpired() {
					return page.Paths.Supporter.InviteExpired.Redirect(w, r, appData)
				}

				if err := memberStore.Create(r.Context(), invite); err != nil {
					return err
				}

				loginSession, err := sesh.Login(sessionStore, r)
				if err != nil {
					return page.Paths.Supporter.Start.Redirect(w, r, appData)
				}

				loginSession.OrganisationID = invite.OrganisationID
				loginSession.OrganisationName = invite.OrganisationName

				if err := sesh.SetLoginSession(sessionStore, r, w, loginSession); err != nil {
					return err
				}

				return page.Paths.Supporter.Dashboard.Redirect(w, r, appData)
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
