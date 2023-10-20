package fixtures

import (
	"encoding/base64"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func Donor(
	tmpl template.Template,
	sessionStore sesh.Store,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
) page.Handler {
	progressValues := []string{
		"provideYourDetails",
		"chooseYourAttorneys",
		"chooseYourReplacementAttorneys",
		"chooseWhenTheLpaCanBeUsed",
		"addRestrictionsToTheLpa",
		"chooseYourCertificateProvider",
		"peopleToNotifyAboutYourLpa",
		"checkAndSendToYourCertificateProvider",
		"payForTheLpa",
		"confirmYourIdentity",
		"signTheLpa",
		"signedByCertificateProvider",
		"signedByAttorneys",
		"submitted",
		"registered",
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			lpaType              = r.FormValue("lpa-type")
			progress             = slices.Index(progressValues, r.FormValue("progress"))
			redirect             = r.FormValue("redirect")
			donor                = r.FormValue("donor")
			certificateProvider  = r.FormValue("certificateProvider")
			attorneys            = r.FormValue("attorneys")
			peopleToNotify       = r.FormValue("peopleToNotify")
			replacementAttorneys = r.FormValue("replacementAttorneys")
			feeType              = r.FormValue("feeType")
		)

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData})
		}

		var (
			donorSub       = random.String(16)
			donorSessionID = base64.StdEncoding.EncodeToString([]byte(donorSub))
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: donorSub, Email: testEmail}); err != nil {
			return err
		}

		lpa, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "provideYourDetails") {
			lpa.Donor = makeDonor()
			lpa.Type = page.LpaTypePropertyFinance
			if lpaType == "hw" {
				lpa.Type = page.LpaTypeHealthWelfare
			}
			lpa.UID = random.UuidString()

			if donor == "cannot-sign" {
				lpa.Donor.ThinksCanSign = actor.No
				lpa.Donor.CanSign = form.No

				lpa.AuthorisedSignatory = actor.AuthorisedSignatory{
					FirstNames: "Allie",
					LastName:   "Adams",
				}

				lpa.IndependentWitness = actor.IndependentWitness{
					FirstNames: "Indie",
					LastName:   "Irwin",
				}
			}

			lpa.Tasks.YourDetails = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "chooseYourAttorneys") {
			lpa.Attorneys.Attorneys = []actor.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])}
			lpa.AttorneyDecisions.How = actor.JointlyAndSeverally

			switch attorneys {
			case "without-address":
				lpa.Attorneys.Attorneys[1].ID = "without-address"
				lpa.Attorneys.Attorneys[1].Address = place.Address{}
			case "trust-corporation-without-address":
				lpa.Attorneys.TrustCorporation = makeTrustCorporation("First Choice Trust Corporation Ltd.")
				lpa.Attorneys.TrustCorporation.Address = place.Address{}
			case "single":
				lpa.Attorneys.Attorneys = lpa.Attorneys.Attorneys[:1]
				lpa.AttorneyDecisions = actor.AttorneyDecisions{}
			case "jointly":
				lpa.AttorneyDecisions.How = actor.Jointly
			case "mixed":
				lpa.AttorneyDecisions.How = actor.JointlyForSomeSeverallyForOthers
				lpa.AttorneyDecisions.Details = "do this and that"
			}

			lpa.Tasks.ChooseAttorneys = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "chooseYourReplacementAttorneys") {
			lpa.ReplacementAttorneys.Attorneys = []actor.Attorney{makeAttorney(replacementAttorneyNames[0]), makeAttorney(replacementAttorneyNames[1])}
			lpa.ReplacementAttorneyDecisions.How = actor.JointlyAndSeverally

			switch replacementAttorneys {
			case "without-address":
				lpa.ReplacementAttorneys.Attorneys[1].ID = "without-address"
				lpa.ReplacementAttorneys.Attorneys[1].Address = place.Address{}
			case "trust-corporation-without-address":
				lpa.ReplacementAttorneys.TrustCorporation = makeTrustCorporation("First Choice Trust Corporation Ltd.")
				lpa.ReplacementAttorneys.TrustCorporation.Address = place.Address{}
			case "single":
				lpa.ReplacementAttorneys.Attorneys = lpa.ReplacementAttorneys.Attorneys[:1]
				lpa.ReplacementAttorneyDecisions = actor.AttorneyDecisions{}
			}

			lpa.Tasks.ChooseReplacementAttorneys = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "chooseWhenTheLpaCanBeUsed") {
			if lpa.Type == page.LpaTypeHealthWelfare {
				lpa.LifeSustainingTreatmentOption = page.LifeSustainingTreatmentOptionA
				lpa.Tasks.LifeSustainingTreatment = actor.TaskCompleted
			} else {
				lpa.WhenCanTheLpaBeUsed = page.CanBeUsedWhenHasCapacity
				lpa.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
			}
		}

		if progress >= slices.Index(progressValues, "addRestrictionsToTheLpa") {
			lpa.Restrictions = "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently"
			lpa.Tasks.Restrictions = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "chooseYourCertificateProvider") {
			lpa.CertificateProvider = makeCertificateProvider()
			if certificateProvider == "paper" {
				lpa.CertificateProvider.CarryOutBy = actor.Paper
			}

			lpa.Tasks.CertificateProvider = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "peopleToNotifyAboutYourLpa") {
			lpa.DoYouWantToNotifyPeople = form.Yes
			lpa.PeopleToNotify = []actor.PersonToNotify{makePersonToNotify(peopleToNotifyNames[0]), makePersonToNotify(peopleToNotifyNames[1])}
			switch peopleToNotify {
			case "without-address":
				lpa.PeopleToNotify[0].ID = "without-address"
				lpa.PeopleToNotify[0].Address = place.Address{}
			case "max":
				lpa.PeopleToNotify = append(lpa.PeopleToNotify, makePersonToNotify(peopleToNotifyNames[2]), makePersonToNotify(peopleToNotifyNames[3]), makePersonToNotify(peopleToNotifyNames[4]))
			}

			lpa.Tasks.PeopleToNotify = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "checkAndSendToYourCertificateProvider") {
			lpa.CheckedAndHappy = true
			lpa.Tasks.CheckYourLpa = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "payForTheLpa") {
			if feeType == "half-fee" {
				lpa.FeeType = page.HalfFee
				lpa.Evidence = page.Evidence{Documents: []page.Document{
					{Key: "evidence-key", Filename: "supporting-evidence.png", Sent: time.Now()},
				}}
			} else {
				lpa.FeeType = page.FullFee
			}

			lpa.PaymentDetails = append(lpa.PaymentDetails, page.Payment{
				PaymentReference: random.String(12),
				PaymentId:        random.String(12),
			})
			lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
		}

		if progress >= slices.Index(progressValues, "confirmYourIdentity") {
			lpa.DonorIdentityUserData = identity.UserData{
				OK:          true,
				RetrievedAt: time.Now(),
				FirstNames:  lpa.Donor.FirstNames,
				LastName:    lpa.Donor.LastName,
				DateOfBirth: lpa.Donor.DateOfBirth,
			}
			lpa.Tasks.ConfirmYourIdentityAndSign = actor.TaskInProgress
		}

		if progress >= slices.Index(progressValues, "signTheLpa") {
			lpa.WantToApplyForLpa = true
			lpa.WantToSignLpa = true
			lpa.SignedAt = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
			lpa.WitnessedByCertificateProviderAt = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
			lpa.Tasks.ConfirmYourIdentityAndSign = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "signedByCertificateProvider") {
			ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: lpa.ID})

			certificateProvider, err := certificateProviderStore.Create(ctx, donorSessionID)
			if err != nil {
				return err
			}

			certificateProvider.Certificate = actor.Certificate{Agreed: time.Now()}

			if err := certificateProviderStore.Put(ctx, certificateProvider); err != nil {
				return err
			}
		}

		if progress >= slices.Index(progressValues, "signedByAttorneys") {
			for isReplacement, list := range map[bool]actor.Attorneys{false: lpa.Attorneys, true: lpa.ReplacementAttorneys} {
				for _, a := range list.Attorneys {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: lpa.ID})

					attorney, err := attorneyStore.Create(ctx, donorSessionID, a.ID, isReplacement, false)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.LpaSignedAt = lpa.SignedAt
					attorney.Confirmed = lpa.SignedAt.Add(2 * time.Hour)

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}
				}

				if list.TrustCorporation.Name != "" {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: lpa.ID})

					attorney, err := attorneyStore.Create(ctx, donorSessionID, "", isReplacement, true)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.WouldLikeSecondSignatory = form.No
					attorney.AuthorisedSignatories = [2]actor.TrustCorporationSignatory{{
						FirstNames:        "A",
						LastName:          "Sign",
						ProfessionalTitle: "Assistant to the signer",
						LpaSignedAt:       lpa.SignedAt,
						Confirmed:         lpa.SignedAt.Add(2 * time.Hour),
					}}

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}
				}
			}
		}

		if progress >= slices.Index(progressValues, "submitted") {
			lpa.SubmittedAt = time.Now()
		}

		if progress >= slices.Index(progressValues, "registered") {
			lpa.RegisteredAt = time.Now()
		}

		donorCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: lpa.ID})

		if err := donorStore.Put(donorCtx, lpa); err != nil {
			return err
		}

		if redirect == "" {
			redirect = page.Paths.Dashboard.Format()
		} else {
			redirect = "/lpa/" + lpa.ID + redirect
		}

		random.UseTestCode = true
		return page.AppData{}.Redirect(w, r, nil, redirect)
	}
}
