package fixtures

import (
	"encoding/base64"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func Dashboard(
	tmpl template.Template,
	sessionStore *sesh.Store,
	donorStore page.DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	shareCodeStore ShareCodeStore,
) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

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

		if err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: meSub, Email: testEmail}); err != nil {
			return err
		}

		if asDonor {
			donor, err := donorStore.Create(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: meSessionID}))
			if err != nil {
				return err
			}

			donorCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: meSessionID, LpaID: donor.LpaID})

			donor.LpaUID = makeUID()
			donor.Donor = makeDonor(testEmail)
			donor.Type = lpadata.LpaTypePropertyAndAffairs

			donor.Attorneys = donordata.Attorneys{
				Attorneys: []donordata.Attorney{makeAttorney(attorneyNames[0])},
			}

			if err := donorStore.Put(donorCtx, donor); err != nil {
				return err
			}
		}

		if asCertificateProvider {
			donor, err := donorStore.Create(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}))
			if err != nil {
				return err
			}
			donor.Donor = makeDonor(testEmail)
			donor.LpaUID = makeUID()

			if err := donorStore.Put(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, LpaID: donor.LpaID}), donor); err != nil {
				return err
			}

			certificateProviderCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: meSessionID, LpaID: donor.LpaID})

			_, err = createCertificateProvider(certificateProviderCtx, shareCodeStore, certificateProviderStore, donor.CertificateProvider.UID, donor.SK, testEmail)
			if err != nil {
				return err
			}
		}

		if asAttorney {
			donor, err := donorStore.Create(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}))
			if err != nil {
				return err
			}
			donor.Donor = makeDonor(testEmail)
			donor.Attorneys = donordata.Attorneys{
				Attorneys: []donordata.Attorney{makeAttorney(attorneyNames[0])},
			}
			donor.LpaUID = makeUID()

			if err := donorStore.Put(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, LpaID: donor.LpaID}), donor); err != nil {
				return err
			}

			attorneyCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: meSessionID, LpaID: donor.LpaID})

			attorney, err := createAttorney(
				attorneyCtx,
				shareCodeStore,
				attorneyStore,
				donor.Attorneys.Attorneys[0].UID,
				false,
				false,
				donor.SK,
				donor.Attorneys.Attorneys[0].Email,
			)
			if err != nil {
				return err
			}

			if err := attorneyStore.Put(attorneyCtx, attorney); err != nil {
				return err
			}
		}

		http.Redirect(w, r, page.PathDashboard.Format(), http.StatusFound)
		return nil
	}
}
