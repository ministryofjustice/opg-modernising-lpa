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
			sub       = randomString(16)
			sessionID = base64.StdEncoding.EncodeToString([]byte(sub))
		)

		lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: sessionID}))
		if err != nil {
			logger.Print("creating lpa ", err)
		}

		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: sessionID, LpaID: lpa.ID})

		if r.FormValue("withDonorDetails") != "" || r.FormValue("completeLpa") != "" {
			CompleteDonorDetails(lpa)
		}

		loginSesh := &sesh.LoginSession{Sub: sub, Email: TestEmail}
		_ = sesh.SetLoginSession(store, r, w, loginSesh)

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

			shareCodeSender.SendCertificateProvider(ctx, notify.CertificateProviderInviteEmail, AppData{
				SessionID: sessionID,
				LpaID:     lpa.ID,
				Localizer: localizer,
			}, false, lpa)

			r.Form.Set("redirect", Paths.CertificateProviderStart)
		}

		if r.FormValue("asCertificateProvider") != "" || r.FormValue("provideCertificate") != "" {
			certificateProvider, err := certificateProviderStore.Create(ctx, sessionID)
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

				certificateProvider.Certificate = actor.Certificate{
					AgreeToStatement: true,
					Agreed:           time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				}
			}

			err = certificateProviderStore.Put(ctx, certificateProvider)
			if err != nil {
				logger.Print("provideCertificate putting CP ", err)
			}
		}

		if r.FormValue("withCertificateProvider") != "" {
			certificateProviderCtx := ctx
			certificateProviderSessionID := base64.StdEncoding.EncodeToString([]byte(randomString(16)))

			if r.FormValue("fresh") != "" {
				lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: certificateProviderSessionID}))
				if err != nil {
					logger.Print("creating lpa ", err)
				}
				certificateProviderCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: sessionID, LpaID: lpa.ID})
			}

			certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, certificateProviderSessionID)
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

			err = certificateProviderStore.Put(certificateProviderCtx, certificateProvider)
			if err != nil {
				logger.Print("withCertificateProvider putting CP ", err)
			}
		}

		if r.FormValue("asAttorney") != "" {
			attorneyCtx := ctx
			attorneySessionID := base64.StdEncoding.EncodeToString([]byte(randomString(16)))

			if r.FormValue("fresh") != "" {
				lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: attorneySessionID}))
				if err != nil {
					logger.Print("creating lpa ", err)
				}
				attorneyCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: sessionID, LpaID: lpa.ID})
			}

			_, err := attorneyStore.Create(attorneyCtx, attorneySessionID, lpa.Attorneys[0].ID, false)
			if err != nil {
				logger.Print("asAttorney:", err)
			}
		}

		if r.FormValue("asReplacementAttorney") != "" {
			_, err := attorneyStore.Create(ctx, sessionID, lpa.ReplacementAttorneys[0].ID, true)
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

			shareCodeSender.SendAttorneys(ctx, AppData{
				SessionID: sessionID,
				LpaID:     lpa.ID,
				Localizer: localizer,
			}, lpa)
		}

		if r.FormValue("signedByDonor") != "" {
			lpa.Submitted = now()
		}

		err = donorStore.Put(ctx, lpa)
		if err != nil {
			logger.Print("putting lpa ", err)
		}

		random.UseTestCode = true

		AppData{}.Redirect(w, r.WithContext(ctx), lpa, r.FormValue("redirect"))
	}
}
