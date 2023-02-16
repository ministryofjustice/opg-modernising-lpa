package page

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func TestingStart(store sesh.Store, lpaStore LpaStore, randomString func(int) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sub := randomString(12)
		sessionID := base64.StdEncoding.EncodeToString([]byte(sub))

		_ = sesh.SetDonor(store, r, w, &sesh.DonorSession{Sub: sub, Email: "simulate-delivered@notifications.service.gov.uk"})

		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: sessionID})

		lpa, _ := lpaStore.Create(ctx)

		if r.FormValue("withDonorDetails") != "" || r.FormValue("completeLpa") != "" {
			lpa.You = MakePerson()
			lpa.WhoFor = "me"
			lpa.Type = "pfa"
			lpa.Tasks.YourDetails = TaskCompleted
		}

		if r.FormValue("withAttorney") != "" {
			lpa.Attorneys = actor.Attorneys{MakeAttorney("John")}

			lpa.Tasks.ChooseAttorneys = TaskCompleted
		}

		if r.FormValue("withAttorneys") != "" || r.FormValue("completeLpa") != "" {
			lpa.Attorneys = actor.Attorneys{
				MakeAttorney("John"),
				MakeAttorney("Joan"),
			}

			lpa.HowAttorneysMakeDecisions = JointlyAndSeverally
			lpa.Tasks.ChooseAttorneys = TaskCompleted
		}

		if r.FormValue("withIncompleteAttorneys") != "" {
			withAddress := MakeAttorney("John")
			withAddress.ID = "with-address"
			withoutAddress := MakeAttorney("Joan")
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
			switch r.FormValue("howAttorneysAct") {
			case Jointly:
				lpa.HowAttorneysMakeDecisions = Jointly
			case JointlyAndSeverally:
				lpa.HowAttorneysMakeDecisions = JointlyAndSeverally
			default:
				lpa.HowAttorneysMakeDecisions = JointlyForSomeSeverallyForOthers
				lpa.HowAttorneysMakeDecisionsDetails = "some details"
			}
		}

		if r.FormValue("withReplacementAttorneys") != "" || r.FormValue("completeLpa") != "" {
			lpa.ReplacementAttorneys = actor.Attorneys{
				MakeAttorney("Jane"),
				MakeAttorney("Jorge"),
			}
			lpa.WantReplacementAttorneys = "yes"
			lpa.HowReplacementAttorneysMakeDecisions = JointlyAndSeverally
			lpa.HowShouldReplacementAttorneysStepIn = OneCanNoLongerAct
			lpa.Tasks.ChooseReplacementAttorneys = TaskCompleted
		}

		if r.FormValue("whenCanBeUsedComplete") != "" || r.FormValue("completeLpa") != "" {
			lpa.WhenCanTheLpaBeUsed = UsedWhenRegistered
			lpa.Tasks.WhenCanTheLpaBeUsed = TaskCompleted
		}

		if r.FormValue("withRestrictions") != "" || r.FormValue("completeLpa") != "" {
			lpa.Restrictions = "Some restrictions on how Attorneys act"
			lpa.Tasks.Restrictions = TaskCompleted
		}

		if r.FormValue("withCP") == "1" || r.FormValue("completeLpa") != "" {
			lpa.CertificateProvider = MakeCertificateProvider("Barbara")
			lpa.Tasks.CertificateProvider = TaskCompleted
		}

		if r.FormValue("withPeopleToNotify") == "1" || r.FormValue("completeLpa") != "" {
			lpa.PeopleToNotify = actor.PeopleToNotify{
				MakePersonToNotify("Joanna"),
				MakePersonToNotify("Jonathan"),
			}
			lpa.DoYouWantToNotifyPeople = "yes"
			lpa.Tasks.PeopleToNotify = TaskCompleted
		}

		if r.FormValue("withIncompletePeopleToNotify") == "1" {
			joanna := MakePersonToNotify("Joanna")
			joanna.Address = place.Address{}
			lpa.PeopleToNotify = actor.PeopleToNotify{
				joanna,
			}
			lpa.DoYouWantToNotifyPeople = "yes"
		}

		if r.FormValue("lpaChecked") == "1" || r.FormValue("completeLpa") != "" {
			lpa.Checked = true
			lpa.HappyToShare = true
			lpa.Tasks.CheckYourLpa = TaskCompleted
		}

		if r.FormValue("paymentComplete") == "1" {
			sesh.SetPayment(store, r, w, &sesh.PaymentSession{PaymentID: random.String(12)})
			lpa.Tasks.PayForLpa = TaskCompleted
		}

		if r.FormValue("idConfirmedAndSigned") == "1" || r.FormValue("completeLpa") != "" {
			lpa.OneLoginUserData = identity.UserData{
				OK:          true,
				RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				FullName:    "Jose Smith",
			}

			lpa.WantToApplyForLpa = true
			lpa.WantToSignLpa = true
			lpa.Submitted = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
			lpa.CPWitnessCodeValidated = true
			lpa.Tasks.ConfirmYourIdentityAndSign = TaskCompleted

		}

		if r.FormValue("withPayment") == "1" || r.FormValue("completeLpa") != "" {
			lpa.Tasks.PayForLpa = TaskCompleted
		}

		if r.FormValue("cookiesAccepted") == "1" {
			http.SetCookie(w, &http.Cookie{
				Name:   "cookies-consent",
				Value:  "accept",
				MaxAge: 365 * 24 * 60 * 60,
				Path:   "/",
			})
		}

		if r.FormValue("asCertificateProvider") == "1" {
			_ = sesh.SetCertificateProvider(store, r, w, &sesh.CertificateProviderSession{
				Sub:            randomString(12),
				Email:          "simulate-delivered@notifications.service.gov.uk",
				DonorSessionID: sessionID,
				LpaID:          lpa.ID,
			})

			lpa.CertificateProviderUserData.FullName = "Barbara Smith"
			lpa.CertificateProviderUserData.OK = true
		}

		_ = lpaStore.Put(ctx, lpa)

		random.UseTestCode = true

		AppData{}.Redirect(w, r.WithContext(ctx), lpa, r.FormValue("redirect"))
	}
}
