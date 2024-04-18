package fixtures

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type DynamoClient interface {
	OneByUID(ctx context.Context, uid string, v interface{}) error
	Create(ctx context.Context, v interface{}) error
}

type DocumentStore interface {
	GetAll(context.Context) (page.Documents, error)
	Put(context.Context, page.Document) error
	Create(ctx context.Context, donor *actor.DonorProvidedDetails, filename string, data []byte) (page.Document, error)
}

var progressValues = []string{
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

type FixtureData struct {
	LpaType                   string
	Progress                  int
	Redirect                  string
	Donor                     string
	CertificateProvider       string
	Attorneys                 string
	PeopleToNotify            string
	ReplacementAttorneys      string
	FeeType                   string
	PaymentTaskProgress       string
	WithVirus                 bool
	UseRealID                 bool
	CertificateProviderEmail  string
	CertificateProviderMobile string
	DonorSub                  string
}

func Donor(
	tmpl template.Template,
	sessionStore *sesh.Store,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	documentStore DocumentStore,
	eventClient *event.Client,
	lpaStoreClient *lpastore.Client,
) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		data := setFixtureData(r)

		if data.DonorSub == "" {
			data.DonorSub = random.String(16)
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: data.DonorSub})
		}

		donorSessionID := base64.StdEncoding.EncodeToString([]byte(data.DonorSub))

		if err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: data.DonorSub, Email: testEmail}); err != nil {
			return err
		}

		donorDetails, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		var fns []func(context.Context, *lpastore.Client, *lpastore.Lpa) error
		donorDetails, fns, err = updateLPAProgress(data, donorDetails, donorSessionID, r, certificateProviderStore, attorneyStore, documentStore, eventClient)
		if err != nil {
			return err
		}

		donorCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: donorDetails.LpaID})

		if err := donorStore.Put(donorCtx, donorDetails); err != nil {
			return err
		}
		if !donorDetails.SignedAt.IsZero() && donorDetails.LpaUID != "" {
			if err := lpaStoreClient.SendLpa(donorCtx, donorDetails); err != nil {
				return err
			}

			lpa, err := lpaStoreClient.Lpa(donorCtx, donorDetails.LpaUID)
			if err != nil {
				return fmt.Errorf("problem getting lpa: %w", err)
			}

			for _, fn := range fns {
				if err := fn(donorCtx, lpaStoreClient, lpa); err != nil {
					return err
				}
			}
		}

		if data.Redirect == "" {
			data.Redirect = page.Paths.Dashboard.Format()
		} else {
			data.Redirect = "/lpa/" + donorDetails.LpaID + data.Redirect
		}

		page.UseTestWitnessCode = true

		http.Redirect(w, r, data.Redirect, http.StatusFound)
		return nil
	}
}

func updateLPAProgress(
	data FixtureData,
	donorDetails *actor.DonorProvidedDetails,
	donorSessionID string,
	r *http.Request,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	documentStore DocumentStore,
	eventClient *event.Client,
) (*actor.DonorProvidedDetails, []func(context.Context, *lpastore.Client, *lpastore.Lpa) error, error) {
	var fns []func(context.Context, *lpastore.Client, *lpastore.Lpa) error

	if data.Progress >= slices.Index(progressValues, "provideYourDetails") {
		donorDetails.Donor = makeDonor()
		donorDetails.Type = actor.LpaTypePropertyAndAffairs

		if data.LpaType == "personal-welfare" {
			donorDetails.Type = actor.LpaTypePersonalWelfare
			donorDetails.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenCapacityLost
		}

		if data.UseRealID {
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
				return nil, nil, err
			}

			donorDetails.HasSentUidRequestedEvent = true
		} else {
			donorDetails.LpaUID = makeUID()
		}

		if data.Donor == "cannot-sign" {
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

	var withoutAddressUID actoruid.UID
	json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users:without-address"`), &withoutAddressUID)

	if data.Progress >= slices.Index(progressValues, "chooseYourAttorneys") {
		donorDetails.Attorneys.Attorneys = []actor.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])}
		donorDetails.AttorneyDecisions.How = actor.JointlyAndSeverally

		switch data.Attorneys {
		case "without-address":
			donorDetails.Attorneys.Attorneys[1].UID = withoutAddressUID
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

	if data.Progress >= slices.Index(progressValues, "chooseYourReplacementAttorneys") {
		donorDetails.ReplacementAttorneys.Attorneys = []actor.Attorney{makeAttorney(replacementAttorneyNames[0]), makeAttorney(replacementAttorneyNames[1])}
		donorDetails.HowShouldReplacementAttorneysStepIn = actor.ReplacementAttorneysStepInWhenOneCanNoLongerAct

		switch data.ReplacementAttorneys {
		case "without-address":
			donorDetails.ReplacementAttorneys.Attorneys[1].UID = withoutAddressUID
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

	if data.Progress >= slices.Index(progressValues, "chooseWhenTheLpaCanBeUsed") {
		if donorDetails.Type == actor.LpaTypePersonalWelfare {
			donorDetails.LifeSustainingTreatmentOption = actor.LifeSustainingTreatmentOptionA
			donorDetails.Tasks.LifeSustainingTreatment = actor.TaskCompleted
		} else {
			donorDetails.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenHasCapacity
			donorDetails.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
		}
	}

	if data.Progress >= slices.Index(progressValues, "addRestrictionsToTheLpa") {
		donorDetails.Restrictions = "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently"
		donorDetails.Tasks.Restrictions = actor.TaskCompleted
	}

	if data.Progress >= slices.Index(progressValues, "chooseYourCertificateProvider") {
		donorDetails.CertificateProvider = makeCertificateProvider()
		if data.CertificateProvider == "paper" {
			donorDetails.CertificateProvider.CarryOutBy = actor.Paper
		}

		if data.CertificateProviderEmail != "" {
			donorDetails.CertificateProvider.Email = data.CertificateProviderEmail
		}

		if data.CertificateProviderMobile != "" {
			donorDetails.CertificateProvider.Mobile = data.CertificateProviderMobile
		}

		donorDetails.Tasks.CertificateProvider = actor.TaskCompleted
	}

	if data.Progress >= slices.Index(progressValues, "peopleToNotifyAboutYourLpa") {
		donorDetails.DoYouWantToNotifyPeople = form.Yes
		donorDetails.PeopleToNotify = []actor.PersonToNotify{makePersonToNotify(peopleToNotifyNames[0]), makePersonToNotify(peopleToNotifyNames[1])}
		switch data.PeopleToNotify {
		case "without-address":
			donorDetails.PeopleToNotify[0].UID = withoutAddressUID
			donorDetails.PeopleToNotify[0].Address = place.Address{}
		case "max":
			donorDetails.PeopleToNotify = append(donorDetails.PeopleToNotify, makePersonToNotify(peopleToNotifyNames[2]), makePersonToNotify(peopleToNotifyNames[3]), makePersonToNotify(peopleToNotifyNames[4]))
		}

		donorDetails.Tasks.PeopleToNotify = actor.TaskCompleted
	}

	if data.Progress >= slices.Index(progressValues, "checkAndSendToYourCertificateProvider") {
		donorDetails.CheckedAt = time.Now()
		donorDetails.Tasks.CheckYourLpa = actor.TaskCompleted
	}

	if data.Progress >= slices.Index(progressValues, "payForTheLpa") {
		if data.FeeType != "" && data.FeeType != "FullFee" {
			feeType, err := pay.ParseFeeType(data.FeeType)
			if err != nil {
				return nil, nil, err
			}

			donorDetails.FeeType = feeType

			stagedForUpload, err := documentStore.Create(
				page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}),
				donorDetails,
				"supporting-evidence.png",
				make([]byte, 64),
			)

			if err != nil {
				return nil, nil, err
			}

			stagedForUpload.Scanned = true
			stagedForUpload.VirusDetected = data.WithVirus

			if err := documentStore.Put(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}), stagedForUpload); err != nil {
				return nil, nil, err
			}

			previouslyUploaded, err := documentStore.Create(
				page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}),
				donorDetails,
				"previously-uploaded-evidence.png",
				make([]byte, 64),
			)

			if err != nil {
				return nil, nil, err
			}

			previouslyUploaded.Scanned = true
			previouslyUploaded.VirusDetected = false
			previouslyUploaded.Sent = time.Now()

			if err := documentStore.Put(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}), previouslyUploaded); err != nil {
				return nil, nil, err
			}
		} else {
			donorDetails.FeeType = pay.FullFee
		}

		donorDetails.PaymentDetails = append(donorDetails.PaymentDetails, actor.Payment{
			PaymentReference: random.String(12),
			PaymentId:        random.String(12),
		})

		donorDetails.Tasks.PayForLpa = actor.PaymentTaskCompleted

		if data.PaymentTaskProgress != "" {
			taskState, err := actor.ParsePaymentTask(data.PaymentTaskProgress)
			if err != nil {
				return nil, nil, err
			}

			donorDetails.EvidenceDelivery = pay.Upload
			donorDetails.Tasks.PayForLpa = taskState
		}
	}

	if data.Progress >= slices.Index(progressValues, "confirmYourIdentity") {
		donorDetails.DonorIdentityUserData = identity.UserData{
			OK:          true,
			RetrievedAt: time.Now(),
			FirstNames:  donorDetails.Donor.FirstNames,
			LastName:    donorDetails.Donor.LastName,
			DateOfBirth: donorDetails.Donor.DateOfBirth,
		}
		donorDetails.Tasks.ConfirmYourIdentityAndSign = actor.TaskInProgress
	}

	if data.Progress >= slices.Index(progressValues, "signTheLpa") {
		donorDetails.WantToApplyForLpa = true
		donorDetails.WantToSignLpa = true
		donorDetails.SignedAt = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
		donorDetails.WitnessedByCertificateProviderAt = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
		donorDetails.Tasks.ConfirmYourIdentityAndSign = actor.TaskCompleted
	}

	if data.Progress >= slices.Index(progressValues, "signedByCertificateProvider") {
		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donorDetails.LpaID})

		certificateProvider, err := certificateProviderStore.Create(ctx, donorSessionID, donorDetails.CertificateProvider.UID)
		if err != nil {
			return nil, nil, err
		}

		certificateProvider.ContactLanguagePreference = localize.En
		certificateProvider.Certificate = actor.Certificate{Agreed: time.Now()}

		if err := certificateProviderStore.Put(ctx, certificateProvider); err != nil {
			return nil, nil, err
		}

		fns = append(fns, func(ctx context.Context, client *lpastore.Client, _ *lpastore.Lpa) error {
			return client.SendCertificateProvider(ctx, donorDetails.LpaUID, certificateProvider)
		})
	}

	if data.Progress >= slices.Index(progressValues, "signedByAttorneys") {
		for isReplacement, list := range map[bool]actor.Attorneys{false: donorDetails.Attorneys, true: donorDetails.ReplacementAttorneys} {
			for _, a := range list.Attorneys {
				ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donorDetails.LpaID})

				attorney, err := attorneyStore.Create(ctx, donorSessionID, a.UID, isReplacement, false)
				if err != nil {
					return nil, nil, err
				}

				attorney.Mobile = testMobile
				attorney.ContactLanguagePreference = localize.En
				attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
				attorney.Tasks.ReadTheLpa = actor.TaskCompleted
				attorney.Tasks.SignTheLpa = actor.TaskCompleted
				attorney.Confirmed = donorDetails.SignedAt.Add(2 * time.Hour)

				if err := attorneyStore.Put(ctx, attorney); err != nil {
					return nil, nil, err
				}

				fns = append(fns, func(ctx context.Context, client *lpastore.Client, lpa *lpastore.Lpa) error {
					return client.SendAttorney(ctx, lpa, attorney)
				})
			}

			if list.TrustCorporation.Name != "" {
				ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donorDetails.LpaID})

				attorney, err := attorneyStore.Create(ctx, donorSessionID, list.TrustCorporation.UID, isReplacement, true)
				if err != nil {
					return nil, nil, err
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
					Confirmed:         donorDetails.SignedAt.Add(2 * time.Hour),
				}}

				if err := attorneyStore.Put(ctx, attorney); err != nil {
					return nil, nil, err
				}

				fns = append(fns, func(ctx context.Context, client *lpastore.Client, lpa *lpastore.Lpa) error {
					return client.SendAttorney(ctx, lpa, attorney)
				})
			}
		}
	}

	if data.Progress >= slices.Index(progressValues, "submitted") {
		donorDetails.SubmittedAt = time.Now()
	}

	if data.Progress == slices.Index(progressValues, "withdrawn") {
		donorDetails.WithdrawnAt = time.Now()
	}

	if data.Progress >= slices.Index(progressValues, "registered") {
		fns = append(fns, func(ctx context.Context, client *lpastore.Client, _ *lpastore.Lpa) error {
			return client.SendRegister(ctx, donorDetails.LpaUID)
		})
	}

	return donorDetails, fns, nil
}

func setFixtureData(r *http.Request) FixtureData {
	return FixtureData{
		LpaType:                   r.FormValue("lpa-type"),
		Progress:                  slices.Index(progressValues, r.FormValue("progress")),
		Redirect:                  r.FormValue("redirect"),
		Donor:                     r.FormValue("donor"),
		CertificateProvider:       r.FormValue("certificateProvider"),
		Attorneys:                 r.FormValue("attorneys"),
		PeopleToNotify:            r.FormValue("peopleToNotify"),
		ReplacementAttorneys:      r.FormValue("replacementAttorneys"),
		FeeType:                   r.FormValue("feeType"),
		PaymentTaskProgress:       r.FormValue("paymentTaskProgress"),
		WithVirus:                 r.FormValue("withVirus") == "1",
		UseRealID:                 r.FormValue("uid") == "real",
		CertificateProviderEmail:  r.FormValue("certificateProviderEmail"),
		CertificateProviderMobile: r.FormValue("certificateProviderMobile"),
		DonorSub:                  r.FormValue("donorSub"),
	}
}
