package fixtures

import (
	"encoding/base64"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

func CertificateProvider(
	tmpl template.Template,
	sessionStore *sesh.Store,
	shareCodeSender ShareCodeSender,
	donorStore page.DonorStore,
	certificateProviderStore CertificateProviderStore,
	eventClient *event.Client,
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
			certificateProviderSub            = r.FormValue("certificateProviderSub")
			shareCode                         = r.FormValue("withShareCode")
			useRealUID                        = r.FormValue("uid") == "real"
			donorActingOnString               = r.FormValue("donorActingOn")
		)

		if certificateProviderSub == "" {
			certificateProviderSub = random.String(16)
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: certificateProviderSub})
		}

		var (
			donorSub                     = random.String(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
		)

		if err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: certificateProviderSub, Email: testEmail}); err != nil {
			return err
		}

		donorDetails, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		var (
			donorCtx               = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: donorDetails.LpaID})
			certificateProviderCtx = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: certificateProviderSessionID, LpaID: donorDetails.LpaID})
		)

		donorDetails.Donor = makeDonor()
		donorDetails.Type = actor.LpaTypePropertyAndAffairs

		if donorActingOnString == "paper" {
			donorDetails.ActingOn = actor.Paper
		}

		if lpaType == "personal-welfare" {
			donorDetails.Type = actor.LpaTypePersonalWelfare
			donorDetails.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenCapacityLost
			donorDetails.LifeSustainingTreatmentOption = actor.LifeSustainingTreatmentOptionA
		} else {
			donorDetails.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenHasCapacity
		}

		if useRealUID {
			if err := eventClient.SendUidRequested(r.Context(), event.UidRequested{
				LpaID:          donorDetails.LpaID,
				DonorSessionID: donorSessionID,
				Type:           donorDetails.Type.String(),
				Donor: uid.DonorDetails{
					Name:     donorDetails.Donor.FullName(),
					Dob:      donorDetails.Donor.DateOfBirth,
					Postcode: donorDetails.Donor.Address.Postcode,
				},
			}); err != nil {
				return err
			}

			donorDetails.HasSentUidRequestedEvent = true
		} else {
			donorDetails.LpaUID = makeUID()
		}

		donorDetails.Attorneys = actor.Attorneys{
			Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])},
		}

		donorDetails.AttorneyDecisions = actor.AttorneyDecisions{How: actor.JointlyAndSeverally}

		donorDetails.CertificateProvider = makeCertificateProvider()
		if email != "" {
			donorDetails.CertificateProvider.Email = email
		}

		if asProfessionalCertificateProvider {
			donorDetails.CertificateProvider.Relationship = actor.Professionally
		}

		certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID, donorDetails.CertificateProvider.UID, donorDetails.ActingOn)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "paid") {
			donorDetails.PaymentDetails = append(donorDetails.PaymentDetails, actor.Payment{
				PaymentReference: random.String(12),
				PaymentId:        random.String(12),
			})
			donorDetails.Tasks.PayForLpa = actor.PaymentTaskCompleted
		}

		if progress >= slices.Index(progressValues, "signedByDonor") {
			donorDetails.Tasks.ConfirmYourIdentityAndSign = actor.TaskCompleted
			donorDetails.WitnessedByCertificateProviderAt = time.Now()
			donorDetails.SignedAt = time.Now()
		}

		if progress >= slices.Index(progressValues, "confirmYourDetails") {
			certificateProvider.DateOfBirth = date.New("1990", "1", "2")
			certificateProvider.ContactLanguagePreference = localize.En
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
				FirstNames:  donorDetails.CertificateProvider.FirstNames,
				LastName:    donorDetails.CertificateProvider.LastName,
				DateOfBirth: certificateProvider.DateOfBirth,
			}
			certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted
		}

		if err := donorStore.Put(donorCtx, donorDetails); err != nil {
			return err
		}
		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}

		// should only be used in tests as otherwise people can read their emails...
		if shareCode != "" {
			shareCodeSender.UseTestCode(shareCode)
		}

		if email != "" {
			shareCodeSender.SendCertificateProviderInvite(donorCtx, page.AppData{
				SessionID: donorSessionID,
				LpaID:     donorDetails.LpaID,
				Localizer: appData.Localizer,
			}, donorDetails)

			http.Redirect(w, r, page.Paths.CertificateProviderStart.Format(), http.StatusFound)
			return nil
		}

		switch redirect {
		case "":
			redirect = page.Paths.Dashboard.Format()
		case page.Paths.CertificateProviderStart.Format():
			redirect = page.Paths.CertificateProviderStart.Format()
		default:
			redirect = "/certificate-provider/" + donorDetails.LpaID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
