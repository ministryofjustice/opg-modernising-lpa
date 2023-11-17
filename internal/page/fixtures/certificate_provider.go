package fixtures

import (
	"encoding/base64"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func CertificateProvider(
	tmpl template.Template,
	sessionStore sesh.Store,
	shareCodeSender ShareCodeSender,
	donorStore page.DonorStore,
	certificateProviderStore CertificateProviderStore,
) page.Handler {
	progressValues := []string{
		"paid",
		"signedByDonor",
		"confirmYourDetails",
		"confirmYourIdentity",
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			lpaType  = r.FormValue("lpa-type")
			progress = slices.Index(progressValues, r.FormValue("progress"))
			email    = r.FormValue("email")
			redirect = r.FormValue("redirect")
		)

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData})
		}

		var (
			donorSub                     = random.String(16)
			certificateProviderSub       = random.String(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: certificateProviderSub, Email: testEmail}); err != nil {
			return err
		}

		lpa, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		var (
			donorCtx               = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			certificateProviderCtx = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: certificateProviderSessionID, LpaID: lpa.ID})
		)

		lpa.UID = makeUid()
		lpa.Donor = makeDonor()
		lpa.Type = page.LpaTypePropertyFinance
		if lpaType == "hw" {
			lpa.Type = page.LpaTypeHealthWelfare
		}

		lpa.Attorneys = actor.Attorneys{
			Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])},
		}

		lpa.CertificateProvider = makeCertificateProvider()
		if email != "" {
			lpa.CertificateProvider.Email = email
		}

		certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "paid") {
			lpa.PaymentDetails = append(lpa.PaymentDetails, page.Payment{
				PaymentReference: random.String(12),
				PaymentId:        random.String(12),
			})
			lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
		}

		if progress >= slices.Index(progressValues, "signedByDonor") {
			lpa.SignedAt = time.Now()
		}

		if progress >= slices.Index(progressValues, "confirmYourDetails") {
			certificateProvider.DateOfBirth = date.New("1990", "1", "2")
			certificateProvider.Tasks.ConfirmYourDetails = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "confirmYourIdentity") {
			certificateProvider.IdentityUserData = identity.UserData{
				OK:          true,
				RetrievedAt: time.Now(),
				FirstNames:  lpa.CertificateProvider.FirstNames,
				LastName:    lpa.CertificateProvider.LastName,
				DateOfBirth: certificateProvider.DateOfBirth,
			}
			certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted
		}

		if err := donorStore.Put(donorCtx, lpa); err != nil {
			return err
		}
		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}

		// should only be used in tests as otherwise people can read their emails...
		if r.FormValue("use-test-code") == "1" {
			shareCodeSender.UseTestCode()
		}

		if email != "" {
			shareCodeSender.SendCertificateProvider(donorCtx, notify.CertificateProviderInviteEmail, page.AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: appData.Localizer,
			}, true, lpa)

			http.Redirect(w, r, page.Paths.CertificateProviderStart.Format(), http.StatusFound)
			return nil
		}

		if redirect == "" {
			redirect = page.Paths.Dashboard.Format()
		} else {
			redirect = "/certificate-provider/" + lpa.ID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
