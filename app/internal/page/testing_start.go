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

func TestingStart(store sesh.Store, donorStore DonorStore, randomString func(int) string, shareCodeSender shareCodeSender, localizer Localizer, certificateProviderStore CertificateProviderStore, attorneyStore AttorneyStore, logger *logging.Logger, now func() time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			donorSub                     = randomString(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSub       = randomString(16)
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
			attorneySub                  = randomString(16)
			attorneySessionID            = base64.StdEncoding.EncodeToString([]byte(attorneySub))
		)

		lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID}))
		if err != nil {
			logger.Print("creating lpa ", err)
		}

		// These contexts act on the same LPA for different actors
		var (
			donorCtx               = ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			certificateProviderCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: certificateProviderSessionID, LpaID: lpa.ID})
			attorneyCtx            = ContextWithSessionData(r.Context(), &SessionData{SessionID: attorneySessionID, LpaID: lpa.ID})
		)

		if r.FormValue("withDonorDetails") != "" || r.FormValue("completeLpa") != "" {
			CompleteDonorDetails(lpa)
		}

		// loginAs controls which actor we will be pretending to be for the LPA
		switch r.FormValue("loginAs") {
		case "attorney":
			_ = sesh.SetLoginSession(store, r, w, &sesh.LoginSession{Sub: attorneySub, Email: TestEmail})
		case "certificate-provider":
			_ = sesh.SetLoginSession(store, r, w, &sesh.LoginSession{Sub: certificateProviderSub, Email: TestEmail})
		default:
			_ = sesh.SetLoginSession(store, r, w, &sesh.LoginSession{Sub: donorSub, Email: TestEmail})
		}

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

			lpa.Tasks.ChooseAttorneys = actor.TaskInProgress
			lpa.Tasks.ChooseReplacementAttorneys = actor.TaskInProgress
		}

		if r.FormValue("withIncompleteReplacementAttorneys") != "" {
			AddReplacementAttorneys(lpa, 2)
			lpa.ReplacementAttorneys[0].ID = "with-address"
			lpa.ReplacementAttorneys[1].ID = "without-address"
			lpa.ReplacementAttorneys[1].Address = place.Address{}

			lpa.ReplacementAttorneys = lpa.Attorneys
			lpa.WantReplacementAttorneys = "yes"
			lpa.Tasks.ChooseAttorneys = actor.TaskCompleted
			lpa.Tasks.ChooseReplacementAttorneys = actor.TaskInProgress
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

		if r.FormValue("howReplacementAttorneysAct") != "" {
			CompleteHowReplacementAttorneysAct(lpa, r.FormValue("howReplacementAttorneysAct"))
		}

		if r.FormValue("whenCanBeUsedComplete") != "" || r.FormValue("completeLpa") != "" {
			CompleteWhenCanLpaBeUsed(lpa)
		}

		if r.FormValue("withRestrictions") != "" || r.FormValue("completeLpa") != "" {
			CompleteRestrictions(lpa)
		}

		if r.FormValue("withCPDetails") != "" || r.FormValue("completeLpa") != "" {
			AddCertificateProvider(lpa, "Jessie")
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

			lpa.Tasks.PeopleToNotify = actor.TaskInProgress
		}

		if r.FormValue("lpaChecked") != "" || r.FormValue("completeLpa") != "" {
			CompleteCheckYourLpa(lpa)
		}

		if r.FormValue("paymentComplete") != "" || r.FormValue("completeLpa") != "" {
			if r.FormValue("paymentComplete") != "" {
				CompleteSectionOne(lpa)
			}
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

		if r.FormValue("withShareCodeSession") != "" {
			sesh.SetShareCode(store, r, w, &sesh.ShareCodeSession{LpaID: lpa.ID, Identity: false})
		}

		if r.FormValue("startCpFlowDonorHasPaid") != "" || r.FormValue("startCpFlowDonorHasNotPaid") != "" {
			CompleteSectionOne(lpa)

			if r.FormValue("startCpFlowDonorHasPaid") != "" {
				PayForLpa(lpa, store, r, w, randomString(12))
			}

			lpa.CertificateProvider.Email = TestEmail

			if r.FormValue("withEmail") != "" {
				lpa.CertificateProvider.Email = r.FormValue("withEmail")
			}

			shareCodeSender.SendCertificateProvider(donorCtx, notify.CertificateProviderInviteEmail, AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: localizer,
			}, false, lpa)

			r.Form.Set("redirect", Paths.CertificateProviderStart)
		}

		if val := r.FormValue("asCertificateProvider"); val != "" {
			currentCtx := certificateProviderCtx

			// "fresh=1" causes an LPA to be created for the certificate provider (as
			// the donor), we then link this to the donor so they are both each
			// other's certificate provider.
			if r.FormValue("fresh") != "" {
				lpa, err := donorStore.Create(certificateProviderCtx)
				if err != nil {
					logger.Print("creating lpa ", err)
				}
				currentCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			}

			certificateProvider, err := certificateProviderStore.Create(currentCtx, certificateProviderSessionID)
			if err != nil {
				logger.Print("asCertificateProvider creating CP ", err)
			}

			certificateProvider.IdentityUserData = identity.UserData{
				OK:         true,
				Provider:   identity.OneLogin,
				FirstNames: "Jessie",
				LastName:   "Jones",
			}

			if val == "certified" {
				certificateProvider.Mobile = TestMobile
				certificateProvider.Email = TestEmail
				certificateProvider.Certificate = actor.Certificate{
					AgreeToStatement: true,
					Agreed:           time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				}
			}

			if err := certificateProviderStore.Put(currentCtx, certificateProvider); err != nil {
				logger.Print("provideCertificate putting CP ", err)
			}
		}

		if r.FormValue("asAttorney") != "" {
			currentCtx := attorneyCtx
			attorneySessionID := base64.StdEncoding.EncodeToString([]byte(randomString(16)))

			if r.FormValue("fresh") != "" {
				lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: attorneySessionID}))
				if err != nil {
					logger.Print("creating lpa ", err)
				}
				currentCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			}

			_, err := attorneyStore.Create(currentCtx, attorneySessionID, lpa.Attorneys[0].ID, false)
			if err != nil {
				logger.Print("asAttorney:", err)
			}
		}

		if r.FormValue("asReplacementAttorney") != "" {
			_, err := attorneyStore.Create(donorCtx, donorSessionID, lpa.ReplacementAttorneys[0].ID, true)
			if err != nil {
				logger.Print("asReplacementAttorney:", err)
			}
		}

		if r.FormValue("sendAttorneyShare") != "" {
			attorneys := lpa.Attorneys
			if r.FormValue("forReplacementAttorney") != "" {
				attorneys = lpa.ReplacementAttorneys
			}

			if r.FormValue("withEmail") != "" {
				attorneys[0].Email = r.FormValue("withEmail")
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

		err = donorStore.Put(donorCtx, lpa)
		if err != nil {
			logger.Print("putting lpa ", err)
		}

		random.UseTestCode = true

		AppData{}.Redirect(w, r.WithContext(donorCtx), lpa, r.FormValue("redirect"))
	}
}
