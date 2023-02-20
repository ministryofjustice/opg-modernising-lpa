package page

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func TestingStart(store sesh.Store, lpaStore LpaStore, randomString func(int) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sub := randomString(12)
		sessionID := base64.StdEncoding.EncodeToString([]byte(sub))
		donorSesh := &sesh.DonorSession{Sub: sub, Email: "simulate-delivered@notifications.service.gov.uk"}

		_ = sesh.SetDonor(store, r, w, donorSesh)

		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: sessionID})

		lpa, _ := lpaStore.Create(ctx)

		if r.FormValue("withDonorDetails") != "" || r.FormValue("completeLpa") != "" {
			CompleteDonorDetails(lpa)
		}

		if r.FormValue("withAttorney") != "" {
			lpa, _ = AddAttorneys(lpa, 1)
		}

		if r.FormValue("withAttorneys") != "" || r.FormValue("completeLpa") != "" {
			lpa, _ = AddAttorneys(lpa, 2)
		}

		if r.FormValue("withIncompleteAttorneys") != "" {
			var firstNames []string
			lpa, firstNames = AddAttorneys(lpa, 2)

			withAddress, _ := GetAttorneyByFirstNames(lpa, firstNames[0])
			withAddress.ID = "with-address"
			withoutAddress, _ := GetAttorneyByFirstNames(lpa, firstNames[1])
			withoutAddress.ID = "without-address"
			withoutAddress.Address = place.Address{}

			lpa.Attorneys = actor.Attorneys{
				withAddress,
				withoutAddress,
			}

			lpa.ReplacementAttorneys = lpa.Attorneys
			lpa.Type = LpaTypePropertyFinance
			lpa.WhenCanTheLpaBeUsed = UsedWhenRegistered

			lpa.HowAttorneysMakeDecisions = JointlyAndSeverally

			lpa.WantReplacementAttorneys = "yes"
			lpa.HowReplacementAttorneysMakeDecisions = JointlyAndSeverally
			lpa.HowShouldReplacementAttorneysStepIn = OneCanNoLongerAct

			lpa.Tasks.ChooseAttorneys = TaskInProgress
			lpa.Tasks.ChooseReplacementAttorneys = TaskInProgress
		}

		if r.FormValue("howAttorneysAct") != "" {
			lpa = CompleteHowAttorneysAct(lpa, r.FormValue("howAttorneysAct"))
		}

		if r.FormValue("withReplacementAttorneys") != "" || r.FormValue("completeLpa") != "" {
			lpa, _ = AddReplacementAttorneys(lpa, 2)
		}

		if r.FormValue("whenCanBeUsedComplete") != "" || r.FormValue("completeLpa") != "" {
			lpa = CompleteWhenCanLpaBeUsed(lpa)
		}

		if r.FormValue("withRestrictions") != "" || r.FormValue("completeLpa") != "" {
			lpa = CompleteRestrictions(lpa)
		}

		if r.FormValue("withCP") != "" || r.FormValue("completeLpa") != "" {
			lpa = AddCertificateProvider(lpa, "Barbara")
		}

		if r.FormValue("withPeopleToNotify") != "" || r.FormValue("completeLpa") != "" {
			lpa, _ = AddPeopleToNotify(lpa, 2)
		}

		if r.FormValue("withIncompletePeopleToNotify") != "" {
			lpa, _ = AddPeopleToNotify(lpa, 1)

			joanna := lpa.PeopleToNotify[0]
			joanna.Address = place.Address{}
			lpa.PeopleToNotify = actor.PeopleToNotify{
				joanna,
			}

			lpa.Tasks.PeopleToNotify = TaskInProgress
		}

		if r.FormValue("lpaChecked") != "" || r.FormValue("completeLpa") != "" {
			lpa = CompleteCheckYourLpa(lpa)
		}

		if r.FormValue("paymentComplete") != "" || r.FormValue("completeLpa") != "" {
			lpa = PayForLpa(lpa, store, r, w)
		}

		if r.FormValue("idConfirmedAndSigned") != "" || r.FormValue("completeLpa") != "" {
			lpa = ConfirmIdAndSign(lpa)
		}

		if r.FormValue("cookiesAccepted") != "" {
			http.SetCookie(w, &http.Cookie{
				Name:   "cookies-consent",
				Value:  "accept",
				MaxAge: 365 * 24 * 60 * 60,
				Path:   "/",
			})
		}

		if r.FormValue("asCertificateProvider") != "" || r.FormValue("provideCertificate") != "" {
			_ = sesh.SetCertificateProvider(store, r, w, &sesh.CertificateProviderSession{
				Sub:            randomString(12),
				Email:          "simulate-delivered@notifications.service.gov.uk",
				DonorSessionID: sessionID,
				LpaID:          lpa.ID,
			})

			lpa.CertificateProviderUserData.FullName = "Barbara Smith"
			lpa.CertificateProviderUserData.OK = true

		}

		if r.FormValue("provideCertificate") != "" {
			lpa.CertificateProviderProvidedDetails.Mobile = "07535111222"
			lpa.CertificateProviderProvidedDetails.Email = "t@example.org"
			lpa.CertificateProviderProvidedDetails.Address = place.Address{
				Line1:      "5 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			}

			lpa.Certificate = Certificate{
				AgreeToStatement: true,
				Agreed:           time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
			}
		}

		// used to switch back to donor after CP fixtures have run
		if r.FormValue("asDonor") != "" {
			_ = sesh.SetDonor(store, r, w, donorSesh)
		}

		_ = lpaStore.Put(ctx, lpa)

		random.UseTestCode = true

		AppData{}.Redirect(w, r.WithContext(ctx), lpa, r.FormValue("redirect"))
	}
}
