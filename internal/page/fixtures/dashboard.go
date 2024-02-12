package fixtures

import (
	"encoding/base64"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type dashboardFixturesData struct {
	App    page.AppData
	Errors validation.List
}

func Dashboard(
	tmpl template.Template,
	sessionStore sesh.Store,
	shareCodeSender *page.ShareCodeSender,
	donorStore page.DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			asDonor               = r.FormValue("asDonor") == "1"
			asAttorney            = r.FormValue("asAttorney") == "1"
			asCertificateProvider = r.FormValue("asCertificateProvider") == "1"
		)

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData})
		}

		var (
			meSub          = random.String(16)
			donorSub       = random.String(16)
			meSessionID    = base64.StdEncoding.EncodeToString([]byte(meSub))
			donorSessionID = base64.StdEncoding.EncodeToString([]byte(donorSub))
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: meSub, Email: testEmail}); err != nil {
			return err
		}

		if asDonor {
			donor, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: meSessionID}))
			if err != nil {
				return err
			}

			donorCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: meSessionID, LpaID: donor.LpaID})

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

		if asCertificateProvider {
			donor, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
			if err != nil {
				return err
			}
			donor.Donor = makeDonor()
			donor.LpaUID = makeUID()

			if err := donorStore.Put(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: donor.LpaID}), donor); err != nil {
				return err
			}

			certificateProviderCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: meSessionID, LpaID: donor.LpaID})

			certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID)
			if err != nil {
				return err
			}

			if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
				return err
			}
		}

		if asAttorney {
			donor, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
			if err != nil {
				return err
			}
			donor.Donor = makeDonor()
			donor.Attorneys = actor.Attorneys{
				Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0])},
			}
			donor.LpaUID = makeUID()

			if err := donorStore.Put(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: donor.LpaID}), donor); err != nil {
				return err
			}

			attorneyCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: meSessionID, LpaID: donor.LpaID})

			attorney, err := attorneyStore.Create(attorneyCtx, donorSessionID, donor.Attorneys.Attorneys[0].UID, false, false)
			if err != nil {
				return err
			}

			if err := attorneyStore.Put(attorneyCtx, attorney); err != nil {
				return err
			}
		}

		http.Redirect(w, r, page.Paths.Dashboard.Format(), http.StatusFound)
		return nil
	}
}
