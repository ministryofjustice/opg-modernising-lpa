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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
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
			lpaType                           = r.FormValue("lpa-type")
			progress                          = slices.Index(progressValues, r.FormValue("progress"))
			email                             = r.FormValue("email")
			redirect                          = r.FormValue("redirect")
			asProfessionalCertificateProvider = r.FormValue("relationship") == "professional"
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
			donorCtx               = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: lpa.LpaID})
			certificateProviderCtx = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: certificateProviderSessionID, LpaID: lpa.LpaID})
		)

		lpa.LpaUID = makeUid()
		lpa.Donor = makeDonor()
		lpa.Type = actor.LpaTypePropertyFinance
		if lpaType == "hw" {
			lpa.Type = actor.LpaTypeHealthWelfare
			lpa.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenCapacityLost
			lpa.LifeSustainingTreatmentOption = actor.LifeSustainingTreatmentOptionA
		} else {
			lpa.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenHasCapacity
		}

		lpa.Attorneys = actor.Attorneys{
			Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])},
		}

		lpa.CertificateProvider = makeCertificateProvider()
		if email != "" {
			lpa.CertificateProvider.Email = email
		}

		if asProfessionalCertificateProvider {
			lpa.CertificateProvider.Relationship = actor.Professionally
		}

		certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "paid") {
			lpa.PaymentDetails = append(lpa.PaymentDetails, actor.Payment{
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

			if asProfessionalCertificateProvider {
				certificateProvider.HomeAddress = place.Address{
					Line1:      "6 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				}
			}
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
				LpaID:     lpa.LpaID,
				Localizer: appData.Localizer,
			}, true, lpa)

			http.Redirect(w, r, page.Paths.CertificateProviderStart.Format(), http.StatusFound)
			return nil
		}

		switch redirect {
		case "":
			redirect = page.Paths.Dashboard.Format()
		case page.Paths.CertificateProviderStart.Format():
			redirect = page.Paths.CertificateProviderStart.Format()
		default:
			redirect = "/certificate-provider/" + lpa.LpaID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
