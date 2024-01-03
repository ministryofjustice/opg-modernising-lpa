package fixtures

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type DynamoClient interface {
	OneByUID(ctx context.Context, uid string, v interface{}) error
	AllByPartialSk(ctx context.Context, pk, partialSk string, v interface{}) error
}

type DocumentStore interface {
	GetAll(context.Context) (page.Documents, error)
	Put(context.Context, page.Document) error
	Create(ctx context.Context, donor *actor.DonorProvidedDetails, filename string, data []byte) (page.Document, error)
}

func Donor(
	tmpl template.Template,
	sessionStore sesh.Store,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	documentStore DocumentStore,
	eventClient *event.Client,
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
		"withdrawn",
		"registered",
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			lpaType                   = r.FormValue("lpa-type")
			progress                  = slices.Index(progressValues, r.FormValue("progress"))
			redirect                  = r.FormValue("redirect")
			donor                     = r.FormValue("donor")
			certificateProvider       = r.FormValue("certificateProvider")
			attorneys                 = r.FormValue("attorneys")
			peopleToNotify            = r.FormValue("peopleToNotify")
			replacementAttorneys      = r.FormValue("replacementAttorneys")
			feeType                   = r.FormValue("feeType")
			paymentTaskProgress       = r.FormValue("paymentTaskProgress")
			withVirus                 = r.FormValue("withVirus") == "1"
			useRealUID                = r.FormValue("uid") == "real"
			certificateProviderEmail  = r.FormValue("certificateProviderEmail")
			certificateProviderMobile = r.FormValue("certificateProviderMobile")
			donorSub                  = r.FormValue("donorSub")
		)

		if donorSub == "" {
			donorSub = random.String(16)
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: donorSub})
		}

		donorSessionID := base64.StdEncoding.EncodeToString([]byte(donorSub))

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: donorSub, Email: testEmail}); err != nil {
			return err
		}

		donorDetails, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "provideYourDetails") {
			donorDetails.Donor = makeDonor()
			donorDetails.Type = actor.LpaTypePropertyAndAffairs
			donorDetails.ContactLanguagePreference = localize.En

			if lpaType == "personal-welfare" {
				donorDetails.Type = actor.LpaTypePersonalWelfare
				donorDetails.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenCapacityLost
			}

			if useRealUID {
				if err := eventClient.SendUidRequested(r.Context(), event.UidRequested{
					LpaID:          donorDetails.LpaID,
					DonorSessionID: donorSessionID,
					Type:           donorDetails.Type.LegacyString(),
					Donor: uid.DonorDetails{
						Name:     donorDetails.Donor.FullName(),
						Dob:      donorDetails.Donor.DateOfBirth,
						Postcode: donorDetails.Donor.Address.Postcode,
					},
				}); err != nil {
					return err
				}
			} else {
				donorDetails.LpaUID = makeUID()
			}

			if donor == "cannot-sign" {
				donorDetails.Donor.ThinksCanSign = actor.No
				donorDetails.Donor.CanSign = form.No

				donorDetails.AuthorisedSignatory = actor.AuthorisedSignatory{
					FirstNames: "Allie",
					LastName:   "Adams",
				}

				donorDetails.IndependentWitness = actor.IndependentWitness{
					FirstNames: "Indie",
					LastName:   "Irwin",
				}
			}

			donorDetails.Tasks.YourDetails = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "chooseYourAttorneys") {
			donorDetails.Attorneys.Attorneys = []actor.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])}
			donorDetails.AttorneyDecisions.How = actor.JointlyAndSeverally

			switch attorneys {
			case "without-address":
				donorDetails.Attorneys.Attorneys[1].ID = "without-address"
				donorDetails.Attorneys.Attorneys[1].Address = place.Address{}
			case "trust-corporation-without-address":
				donorDetails.Attorneys.TrustCorporation = makeTrustCorporation("First Choice Trust Corporation Ltd.")
				donorDetails.Attorneys.TrustCorporation.Address = place.Address{}
			case "trust-corporation":
				donorDetails.Attorneys.TrustCorporation = makeTrustCorporation("First Choice Trust Corporation Ltd.")
			case "single":
				donorDetails.Attorneys.Attorneys = donorDetails.Attorneys.Attorneys[:1]
				donorDetails.AttorneyDecisions = actor.AttorneyDecisions{}
			case "jointly":
				donorDetails.AttorneyDecisions.How = actor.Jointly
			case "jointly-for-some-severally-for-others":
				donorDetails.AttorneyDecisions.How = actor.JointlyForSomeSeverallyForOthers
				donorDetails.AttorneyDecisions.Details = "do this and that"
			}

			donorDetails.Tasks.ChooseAttorneys = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "chooseYourReplacementAttorneys") {
			donorDetails.ReplacementAttorneys.Attorneys = []actor.Attorney{makeAttorney(replacementAttorneyNames[0]), makeAttorney(replacementAttorneyNames[1])}
			donorDetails.HowShouldReplacementAttorneysStepIn = actor.ReplacementAttorneysStepInWhenOneCanNoLongerAct

			switch replacementAttorneys {
			case "without-address":
				donorDetails.ReplacementAttorneys.Attorneys[1].ID = "without-address"
				donorDetails.ReplacementAttorneys.Attorneys[1].Address = place.Address{}
			case "trust-corporation-without-address":
				donorDetails.ReplacementAttorneys.TrustCorporation = makeTrustCorporation("First Choice Trust Corporation Ltd.")
				donorDetails.ReplacementAttorneys.TrustCorporation.Address = place.Address{}
			case "trust-corporation":
				donorDetails.ReplacementAttorneys.TrustCorporation = makeTrustCorporation("First Choice Trust Corporation Ltd.")
			case "single":
				donorDetails.ReplacementAttorneys.Attorneys = donorDetails.ReplacementAttorneys.Attorneys[:1]
				donorDetails.HowShouldReplacementAttorneysStepIn = actor.ReplacementAttorneysStepIn(0)
			}

			donorDetails.Tasks.ChooseReplacementAttorneys = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "chooseWhenTheLpaCanBeUsed") {
			if donorDetails.Type == actor.LpaTypePersonalWelfare {
				donorDetails.LifeSustainingTreatmentOption = actor.LifeSustainingTreatmentOptionA
				donorDetails.Tasks.LifeSustainingTreatment = actor.TaskCompleted
			} else {
				donorDetails.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenHasCapacity
				donorDetails.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
			}
		}

		if progress >= slices.Index(progressValues, "addRestrictionsToTheLpa") {
			donorDetails.Restrictions = "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently"
			donorDetails.Tasks.Restrictions = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "chooseYourCertificateProvider") {
			donorDetails.CertificateProvider = makeCertificateProvider()
			if certificateProvider == "paper" {
				donorDetails.CertificateProvider.CarryOutBy = actor.Paper
			}

			if certificateProviderEmail != "" {
				donorDetails.CertificateProvider.Email = certificateProviderEmail
			}

			if certificateProviderMobile != "" {
				donorDetails.CertificateProvider.Mobile = certificateProviderMobile
			}

			donorDetails.Tasks.CertificateProvider = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "peopleToNotifyAboutYourLpa") {
			donorDetails.DoYouWantToNotifyPeople = form.Yes
			donorDetails.PeopleToNotify = []actor.PersonToNotify{makePersonToNotify(peopleToNotifyNames[0]), makePersonToNotify(peopleToNotifyNames[1])}
			switch peopleToNotify {
			case "without-address":
				donorDetails.PeopleToNotify[0].ID = "without-address"
				donorDetails.PeopleToNotify[0].Address = place.Address{}
			case "max":
				donorDetails.PeopleToNotify = append(donorDetails.PeopleToNotify, makePersonToNotify(peopleToNotifyNames[2]), makePersonToNotify(peopleToNotifyNames[3]), makePersonToNotify(peopleToNotifyNames[4]))
			}

			donorDetails.Tasks.PeopleToNotify = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "checkAndSendToYourCertificateProvider") {
			donorDetails.CheckedAt = time.Now()
			donorDetails.Tasks.CheckYourLpa = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "payForTheLpa") {
			if feeType != "" && feeType != "FullFee" {
				feeType, err := pay.ParseFeeType(feeType)
				if err != nil {
					return err
				}

				donorDetails.FeeType = feeType

				stagedForUpload, err := documentStore.Create(
					page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}),
					donorDetails,
					"supporting-evidence.png",
					make([]byte, 64),
				)

				if err != nil {
					return err
				}

				stagedForUpload.Scanned = true
				stagedForUpload.VirusDetected = withVirus

				if err := documentStore.Put(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}), stagedForUpload); err != nil {
					return err
				}

				previouslyUploaded, err := documentStore.Create(
					page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}),
					donorDetails,
					"previously-uploaded-evidence.png",
					make([]byte, 64),
				)

				if err != nil {
					return err
				}

				previouslyUploaded.Scanned = true
				previouslyUploaded.VirusDetected = false
				previouslyUploaded.Sent = time.Now()

				if err := documentStore.Put(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}), previouslyUploaded); err != nil {
					return err
				}
			} else {
				donorDetails.FeeType = pay.FullFee
			}

			donorDetails.PaymentDetails = append(donorDetails.PaymentDetails, actor.Payment{
				PaymentReference: random.String(12),
				PaymentId:        random.String(12),
			})

			donorDetails.Tasks.PayForLpa = actor.PaymentTaskCompleted

			if paymentTaskProgress != "" {
				taskState, err := actor.ParsePaymentTask(paymentTaskProgress)
				if err != nil {
					return err
				}

				donorDetails.EvidenceDelivery = pay.Upload
				donorDetails.Tasks.PayForLpa = taskState
			}
		}

		if progress >= slices.Index(progressValues, "confirmYourIdentity") {
			donorDetails.DonorIdentityUserData = identity.UserData{
				OK:          true,
				RetrievedAt: time.Now(),
				FirstNames:  donorDetails.Donor.FirstNames,
				LastName:    donorDetails.Donor.LastName,
				DateOfBirth: donorDetails.Donor.DateOfBirth,
			}
			donorDetails.Tasks.ConfirmYourIdentityAndSign = actor.TaskInProgress
		}

		if progress >= slices.Index(progressValues, "signTheLpa") {
			donorDetails.WantToApplyForLpa = true
			donorDetails.WantToSignLpa = true
			donorDetails.SignedAt = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
			donorDetails.WitnessedByCertificateProviderAt = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
			donorDetails.Tasks.ConfirmYourIdentityAndSign = actor.TaskCompleted
		}

		if progress >= slices.Index(progressValues, "signedByCertificateProvider") {
			ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donorDetails.LpaID})

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
			for isReplacement, list := range map[bool]actor.Attorneys{false: donorDetails.Attorneys, true: donorDetails.ReplacementAttorneys} {
				for _, a := range list.Attorneys {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donorDetails.LpaID})

					attorney, err := attorneyStore.Create(ctx, donorSessionID, a.ID, isReplacement, false)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.LpaSignedAt = donorDetails.SignedAt
					attorney.Confirmed = donorDetails.SignedAt.Add(2 * time.Hour)

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}
				}

				if list.TrustCorporation.Name != "" {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donorDetails.LpaID})

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
						LpaSignedAt:       donorDetails.SignedAt,
						Confirmed:         donorDetails.SignedAt.Add(2 * time.Hour),
					}}

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}
				}
			}
		}

		if progress >= slices.Index(progressValues, "submitted") {
			donorDetails.SubmittedAt = time.Now()
		}

		if progress == slices.Index(progressValues, "withdrawn") {
			donorDetails.WithdrawnAt = time.Now()
		}

		if progress >= slices.Index(progressValues, "registered") {
			donorDetails.RegisteredAt = time.Now()
		}

		donorCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: donorDetails.LpaID})

		if err := donorStore.Put(donorCtx, donorDetails); err != nil {
			return err
		}

		if redirect == "" {
			redirect = page.Paths.Dashboard.Format()
		} else {
			redirect = "/lpa/" + donorDetails.LpaID + redirect
		}

		log.Println("Logging in with sub", donorSub)
		random.UseTestCode = true
		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
