package page

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func TestingStart(store sesh.Store, lpaStore LpaStore, randomString func(int) string, shareCodeSender shareCodeSender, localizer Localizer, certificateProviderStore CertificateProviderStore, logger *logging.Logger, now func() time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		donorSub := randomString(16)
		donorSessionID := base64.StdEncoding.EncodeToString([]byte(donorSub))

		cpSub := randomString(16)
		cpSessionID := base64.StdEncoding.EncodeToString([]byte(cpSub))

		lpa, err := lpaStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID}))
		if err != nil {
			logger.Print("creating lpa ", err)
		}

		donorCtx := ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})

		cpCtx := ContextWithSessionData(r.Context(), &SessionData{SessionID: cpSessionID, LpaID: lpa.ID})

		if r.FormValue("withDonorDetails") != "" || r.FormValue("completeLpa") != "" {
			CompleteDonorDetails(lpa)
		}

		donorSesh := &sesh.DonorSession{Sub: donorSub, Email: TestEmail}
		_ = sesh.SetDonor(store, r, w, donorSesh)

		if t := r.FormValue("withType"); t != "" {
			lpa.Type = t
		}

		if r.FormValue("withAttorney") != "" {
			AddAttorneys(lpa, 1)
		}

		if r.FormValue("withAttorneys") != "" || r.FormValue("completeLpa") != "" {
			AddAttorneys(lpa, 2)
		}

		if r.FormValue("withIncompleteAttorneys") != "" {
			firstNames := AddAttorneys(lpa, 2)

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

			lpa.AttorneyDecisions.How = actor.JointlyAndSeverally

			lpa.WantReplacementAttorneys = "yes"
			lpa.ReplacementAttorneyDecisions.How = actor.JointlyAndSeverally
			lpa.HowShouldReplacementAttorneysStepIn = OneCanNoLongerAct

			lpa.Tasks.ChooseAttorneys = TaskInProgress
			lpa.Tasks.ChooseReplacementAttorneys = TaskInProgress
		}

		if r.FormValue("howAttorneysAct") != "" {
			CompleteHowAttorneysAct(lpa, r.FormValue("howAttorneysAct"))
		}

		if r.FormValue("withReplacementAttorney") != "" {
			AddReplacementAttorneys(lpa, 1)
		}

		if r.FormValue("withReplacementAttorneys") != "" || r.FormValue("completeLpa") != "" {
			AddReplacementAttorneys(lpa, 2)
		}

		if r.FormValue("whenCanBeUsedComplete") != "" || r.FormValue("completeLpa") != "" {
			CompleteWhenCanLpaBeUsed(lpa)
		}

		if r.FormValue("withRestrictions") != "" || r.FormValue("completeLpa") != "" {
			CompleteRestrictions(lpa)
		}

		if r.FormValue("withCPDetails") != "" || r.FormValue("completeLpa") != "" {
			AddCertificateProviderDetails(lpa, "Jessie")
		}

		if r.FormValue("withPeopleToNotify") != "" || r.FormValue("completeLpa") != "" {
			count, err := strconv.Atoi(r.FormValue("withPeopleToNotify"))
			if err != nil {
				count = 2
			}

			AddPeopleToNotify(lpa, count)
		}

		if r.FormValue("withIncompletePeopleToNotify") != "" {
			AddPeopleToNotify(lpa, 1)

			joanna := lpa.PeopleToNotify[0]
			joanna.Address = place.Address{}
			lpa.PeopleToNotify = actor.PeopleToNotify{
				joanna,
			}

			lpa.Tasks.PeopleToNotify = TaskInProgress
		}

		if r.FormValue("lpaChecked") != "" || r.FormValue("completeLpa") != "" {
			CompleteCheckYourLpa(lpa)
		}

		if r.FormValue("paymentComplete") != "" || r.FormValue("completeLpa") != "" {
			PayForLpa(lpa, store, r, w, randomString(12))
		}

		if r.FormValue("idConfirmedAndSigned") != "" || r.FormValue("completeLpa") != "" {
			ConfirmIdAndSign(lpa)
		}

		if r.FormValue("cookiesAccepted") != "" {
			http.SetCookie(w, &http.Cookie{
				Name:   "cookies-consent",
				Value:  "accept",
				MaxAge: 365 * 24 * 60 * 60,
				Path:   "/",
			})
		}

		if r.FormValue("useTestShareCode") != "" {
			shareCodeSender.UseTestCode()
		}

		if r.FormValue("startCpFlowDonorHasPaid") != "" || r.FormValue("startCpFlowDonorHasNotPaid") != "" {
			CompleteSectionOne(lpa)

			if r.FormValue("startCpFlowDonorHasPaid") != "" {
				PayForLpa(lpa, store, r, w, randomString(12))
			}

			lpa.CertificateProviderDetails.Email = TestEmail

			if r.FormValue("withEmail") != "" {
				lpa.CertificateProviderDetails.Email = r.FormValue("withEmail")
			}

			shareCodeSender.SendCertificateProvider(donorCtx, notify.CertificateProviderInviteEmail, AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: localizer,
			}, false, lpa)

			r.Form.Set("redirect", Paths.CertificateProviderStart)
		}

		if r.FormValue("asCertificateProvider") != "" || r.FormValue("provideCertificate") != "" {
			_ = sesh.SetCertificateProvider(store, r, w, &sesh.CertificateProviderSession{
				Sub:   cpSub,
				Email: TestEmail,
				LpaID: lpa.ID,
			})

			certificateProvider, err := certificateProviderStore.Create(cpCtx)
			if err != nil {
				logger.Print("asCertificateProvider||provideCertificate creating CP ", err)
			}

			certificateProvider.IdentityUserData = identity.UserData{
				OK:         true,
				Provider:   identity.OneLogin,
				FirstNames: "Jessie",
				LastName:   "Jones",
			}

			if r.FormValue("provideCertificate") != "" {
				certificateProvider.Mobile = TestMobile
				certificateProvider.Email = TestEmail
				certificateProvider.Address = place.Address{
					Line1:      "5 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				}

				certificateProvider.Certificate = actor.Certificate{
					AgreeToStatement: true,
					Agreed:           time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				}
			}

			err = certificateProviderStore.Put(cpCtx, certificateProvider)
			if err != nil {
				logger.Print("provideCertificate putting CP ", err)
			}
		}

		if r.FormValue("withCertificateProvider") != "" {
			certificateProvider, err := certificateProviderStore.Create(cpCtx)

			if err != nil {
				logger.Print("withCertificateProvider creating CP ", err)
			}

			certificateProvider.IdentityUserData = identity.UserData{
				OK:         true,
				Provider:   identity.OneLogin,
				FirstNames: "Jessie",
				LastName:   "Jones",
			}

			certificateProvider.Mobile = TestMobile
			certificateProvider.Email = TestEmail
			certificateProvider.Address = place.Address{
				Line1:      "5 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			}

			err = certificateProviderStore.Put(cpCtx, certificateProvider)
			if err != nil {
				logger.Print("withCertificateProvider putting CP ", err)
			}

			_ = sesh.SetDonor(store, r, w, donorSesh)
		}

		if r.FormValue("asAttorney") != "" {
			_ = sesh.SetAttorney(store, r, w, &sesh.AttorneySession{
				Sub:        randomString(12),
				Email:      TestEmail,
				AttorneyID: lpa.Attorneys[0].ID,
				LpaID:      lpa.ID,
			})
		}

		if r.FormValue("asReplacementAttorney") != "" {
			_ = sesh.SetAttorney(store, r, w, &sesh.AttorneySession{
				Sub:                   randomString(12),
				Email:                 TestEmail,
				AttorneyID:            lpa.ReplacementAttorneys[0].ID,
				LpaID:                 lpa.ID,
				IsReplacementAttorney: true,
			})
		}

		if r.FormValue("sendAttorneyShare") != "" {
			attorneys := actor.Attorneys{MakeAttorney(AttorneyNames[0])}
			attorneys[0].Email = TestEmail

			if r.FormValue("withEmail") != "" {
				attorneys[0].Email = r.FormValue("withEmail")
			}

			if r.FormValue("forReplacementAttorney") != "" {
				lpa.ReplacementAttorneys = attorneys
			} else {
				lpa.Attorneys = attorneys
			}

			shareCodeSender.SendAttorneys(donorCtx, AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: localizer,
			}, lpa)
		}

		if r.FormValue("signedByDonor") != "" {
			lpa.Submitted = now()
		}

		// used to switch back to donor after CP fixtures have run
		if r.FormValue("asDonor") != "" {
			_ = sesh.SetDonor(store, r, w, donorSesh)
		}

		err = lpaStore.Put(donorCtx, lpa)
		if err != nil {
			logger.Print("putting lpa ", err)
		}

		random.UseTestCode = true

		if r.FormValue("asCertificateProvider") != "" || r.FormValue("provideCertificate") != "" {
			AppData{}.Redirect(w, r.WithContext(cpCtx), lpa, r.FormValue("redirect"))
		} else {
			AppData{}.Redirect(w, r.WithContext(donorCtx), lpa, r.FormValue("redirect"))
		}
	}
}
