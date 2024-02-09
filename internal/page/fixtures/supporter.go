package fixtures

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type OrganisationStore interface {
	Create(context.Context, string) (*actor.Organisation, error)
	CreateLPA(context.Context) (*actor.DonorProvidedDetails, error)
}

func Supporter(sessionStore sesh.Store, organisationStore OrganisationStore, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			organisation = r.FormValue("organisation")
			lpa          = r.FormValue("lpa")
			redirect     = r.FormValue("redirect")

			supporterSub       = random.String(16)
			supporterSessionID = base64.StdEncoding.EncodeToString([]byte(supporterSub))
			ctx                = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: supporterSessionID})
		)

		loginSession := &sesh.LoginSession{Sub: supporterSub, Email: testEmail}

		if organisation == "1" {
			org, err := organisationStore.Create(ctx, random.String(12))
			if err != nil {
				return err
			}

			loginSession.OrganisationID = org.ID
			loginSession.OrganisationName = org.Name

			if lpa == "1" {
				donor, err := organisationStore.CreateLPA(page.ContextWithSessionData(r.Context(), &page.SessionData{OrganisationID: org.ID}))
				if err != nil {
					return err
				}

				donorCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{OrganisationID: org.ID, LpaID: donor.LpaID})

				donor.LpaUID = makeUID()
				donor.Donor = makeDonor()
				donor.Type = actor.LpaTypePropertyAndAffairs

				donor.Attorneys = actor.Attorneys{
					Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0])},
				}

				if err := donorStore.Put(donorCtx, donor); err != nil {
					return err
				}
			}
		}

		if err := sesh.SetLoginSession(sessionStore, r, w, loginSession); err != nil {
			return err
		}

		if redirect != page.Paths.Supporter.EnterOrganisationName.Format() {
			redirect = "/supporter/" + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
