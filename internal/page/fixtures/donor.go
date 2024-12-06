package fixtures

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/document"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type DynamoClient interface {
	OneByUID(ctx context.Context, uid string, v interface{}) error
	Create(ctx context.Context, v interface{}) error
}

type DocumentStore interface {
	GetAll(context.Context) (document.Documents, error)
	Put(context.Context, document.Document) error
	Create(ctx context.Context, donor *donordata.Provided, filename string, data []byte) (document.Document, error)
}

var progressValues = []string{
	"provideYourDetails",
	"chooseYourAttorneys",
	"chooseYourReplacementAttorneys",
	"chooseWhenTheLpaCanBeUsed",
	"addRestrictionsToTheLpa",
	"chooseYourCertificateProvider",
	"peopleToNotifyAboutYourLpa",
	"addCorrespondent",
	"checkAndSendToYourCertificateProvider",
	"payForTheLpa",
	"confirmYourIdentity",
	"signTheLpa",
	"signedByCertificateProvider",
	"signedByAttorneys",
	"submitted",
	"statutoryWaitingPeriod",
	// end states
	"registered",
	"withdrawn",
	"certificateProviderOptedOut",
	"doNotRegister",
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
	DonorEmail                string
	IdStatus                  string
	Voucher                   string
	FailedVouchAttempts       string
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
	shareCodeStore ShareCodeStore,
) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		data := setFixtureData(r)

		if data.DonorSub == "" {
			data.DonorSub = random.String(16)
		}

		if data.DonorEmail == "" {
			data.DonorEmail = testEmail
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: data.DonorSub, DonorEmail: data.DonorEmail})
		}

		donorSessionID := base64.StdEncoding.EncodeToString([]byte(data.DonorSub))

		if err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: data.DonorSub, Email: data.DonorEmail}); err != nil {
			return err
		}

		donorDetails, err := donorStore.Create(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		var fns []func(context.Context, *lpastore.Client, *lpadata.Lpa) error
		donorDetails, fns, err = updateLPAProgress(data, donorDetails, donorSessionID, r, certificateProviderStore, attorneyStore, documentStore, eventClient, shareCodeStore)
		if err != nil {
			return err
		}

		donorCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, LpaID: donorDetails.LpaID})

		if data.Progress >= slices.Index(progressValues, "checkAndSendToYourCertificateProvider") {
			if err = donorDetails.UpdateCheckedHash(); err != nil {
				return fmt.Errorf("problem updating checkedHash: %w", err)
			}
		}

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
			data.Redirect = page.PathDashboard.Format()
		} else {
			data.Redirect = "/lpa/" + donorDetails.LpaID + data.Redirect
		}

		donor.UseTestWitnessCode = true

		http.Redirect(w, r, data.Redirect, http.StatusFound)
		return nil
	}
}

func updateLPAProgress(
	data FixtureData,
	donorDetails *donordata.Provided,
	donorSessionID string,
	r *http.Request,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	documentStore DocumentStore,
	eventClient *event.Client,
	shareCodeStore ShareCodeStore,
) (*donordata.Provided, []func(context.Context, *lpastore.Client, *lpadata.Lpa) error, error) {
	var fns []func(context.Context, *lpastore.Client, *lpadata.Lpa) error

	if data.Progress >= slices.Index(progressValues, "provideYourDetails") {
		donorDetails.Donor = makeDonor(data.DonorEmail)

		donorDetails.Type = lpadata.LpaTypePropertyAndAffairs

		if data.LpaType == "personal-welfare" {
			donorDetails.Type = lpadata.LpaTypePersonalWelfare
			donorDetails.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenCapacityLost
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
		} else {
			donorDetails.LpaUID = makeUID()
		}

		if data.Donor == "cannot-sign" {
			donorDetails.Donor.ThinksCanSign = donordata.No
			donorDetails.Donor.CanSign = form.No

			donorDetails.AuthorisedSignatory = donordata.AuthorisedSignatory{
				FirstNames: "Allie",
				LastName:   "Adams",
			}

			donorDetails.IndependentWitness = donordata.IndependentWitness{
				FirstNames: "Indie",
				LastName:   "Irwin",
			}

			donorDetails.Tasks.ChooseYourSignatory = task.StateCompleted
		}

		donorDetails.Tasks.YourDetails = task.StateCompleted
	}

	var withoutAddressUID actoruid.UID
	json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users:without-address"`), &withoutAddressUID)

	if data.Progress >= slices.Index(progressValues, "chooseYourAttorneys") {
		donorDetails.Attorneys.Attorneys = []donordata.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])}
		donorDetails.AttorneyDecisions.How = lpadata.JointlyAndSeverally

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
			donorDetails.AttorneyDecisions = donordata.AttorneyDecisions{}
		case "jointly":
			donorDetails.AttorneyDecisions.How = lpadata.Jointly
		case "jointly-for-some-severally-for-others":
			donorDetails.AttorneyDecisions.How = lpadata.JointlyForSomeSeverallyForOthers
			donorDetails.AttorneyDecisions.Details = "do this and that"
		}

		donorDetails.Tasks.ChooseAttorneys = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "chooseYourReplacementAttorneys") {
		donorDetails.ReplacementAttorneys.Attorneys = []donordata.Attorney{makeAttorney(replacementAttorneyNames[0]), makeAttorney(replacementAttorneyNames[1])}
		donorDetails.HowShouldReplacementAttorneysStepIn = lpadata.ReplacementAttorneysStepInWhenOneCanNoLongerAct

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
			donorDetails.HowShouldReplacementAttorneysStepIn = lpadata.ReplacementAttorneysStepIn(0)
		}

		donorDetails.Tasks.ChooseReplacementAttorneys = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "chooseWhenTheLpaCanBeUsed") {
		if donorDetails.Type == lpadata.LpaTypePersonalWelfare {
			donorDetails.LifeSustainingTreatmentOption = lpadata.LifeSustainingTreatmentOptionA
			donorDetails.Tasks.LifeSustainingTreatment = task.StateCompleted
		} else {
			donorDetails.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenHasCapacity
			donorDetails.Tasks.WhenCanTheLpaBeUsed = task.StateCompleted
		}
	}

	if data.Progress >= slices.Index(progressValues, "addRestrictionsToTheLpa") {
		donorDetails.Restrictions = "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently"
		donorDetails.Tasks.Restrictions = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "chooseYourCertificateProvider") {
		donorDetails.CertificateProvider = makeCertificateProvider()
		if data.CertificateProvider == "paper" {
			donorDetails.CertificateProvider.CarryOutBy = lpadata.ChannelPaper
		}

		if data.CertificateProviderEmail != "" {
			donorDetails.CertificateProvider.Email = data.CertificateProviderEmail
		}

		if data.CertificateProviderMobile != "" {
			donorDetails.CertificateProvider.Mobile = data.CertificateProviderMobile
		}

		donorDetails.Tasks.CertificateProvider = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "peopleToNotifyAboutYourLpa") {
		donorDetails.DoYouWantToNotifyPeople = form.Yes
		donorDetails.PeopleToNotify = []donordata.PersonToNotify{makePersonToNotify(peopleToNotifyNames[0]), makePersonToNotify(peopleToNotifyNames[1])}
		switch data.PeopleToNotify {
		case "without-address":
			donorDetails.PeopleToNotify[0].UID = withoutAddressUID
			donorDetails.PeopleToNotify[0].Address = place.Address{}
		case "max":
			donorDetails.PeopleToNotify = append(donorDetails.PeopleToNotify, makePersonToNotify(peopleToNotifyNames[2]), makePersonToNotify(peopleToNotifyNames[3]), makePersonToNotify(peopleToNotifyNames[4]))
		}

		donorDetails.Tasks.PeopleToNotify = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "addCorrespondent") {
		donorDetails.AddCorrespondent = form.Yes
		donorDetails.Correspondent = makeCorrespondent(Name{
			Firstnames: "Jonathan",
			Lastname:   "Ashfurlong",
		})

		donorDetails.Tasks.AddCorrespondent = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "checkAndSendToYourCertificateProvider") {
		donorDetails.CheckedAt = time.Now()
		donorDetails.Tasks.CheckYourLpa = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "payForTheLpa") {
		if data.FeeType != "" && data.FeeType != "FullFee" {
			feeType, err := pay.ParseFeeType(data.FeeType)
			if err != nil {
				return nil, nil, err
			}

			donorDetails.FeeType = feeType

			stagedForUpload, err := documentStore.Create(
				appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}),
				donorDetails,
				"supporting-evidence.png",
				make([]byte, 64),
			)

			if err != nil {
				return nil, nil, err
			}

			stagedForUpload.Scanned = true
			stagedForUpload.VirusDetected = data.WithVirus

			if err := documentStore.Put(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}), stagedForUpload); err != nil {
				return nil, nil, err
			}

			previouslyUploaded, err := documentStore.Create(
				appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}),
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

			if err := documentStore.Put(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}), previouslyUploaded); err != nil {
				return nil, nil, err
			}
		} else {
			donorDetails.FeeType = pay.FullFee
		}

		donorDetails.PaymentDetails = append(donorDetails.PaymentDetails, donordata.Payment{
			PaymentReference: random.String(12),
			PaymentId:        random.String(12),
		})

		donorDetails.Tasks.PayForLpa = task.PaymentStateCompleted

		if data.PaymentTaskProgress != "" {
			taskState, err := task.ParsePaymentState(data.PaymentTaskProgress)
			if err != nil {
				return nil, nil, err
			}

			donorDetails.EvidenceDelivery = pay.Upload
			donorDetails.Tasks.PayForLpa = taskState
		}
	}

	if data.Progress >= slices.Index(progressValues, "confirmYourIdentity") {
		var userData identity.UserData

		idActor, idStatus, ok := strings.Cut(data.IdStatus, ":")
		if !ok && data.IdStatus != "" {
			return nil, nil, errors.New("invalid value for idStatus - must be in format actor:status")
		}

		switch idStatus {
		case "failed":
			userData = identity.UserData{
				Status: identity.StatusFailed,
			}
		case "insufficient-evidence":
			userData = identity.UserData{
				Status: identity.StatusInsufficientEvidence,
			}
		case "expired":
			userData = identity.UserData{
				Status: identity.StatusExpired,
			}
		case "post-office":
			userData = identity.UserData{}
			donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStatePending
		default:
			userData = identity.UserData{
				Status:      identity.StatusConfirmed,
				CheckedAt:   time.Now(),
				FirstNames:  donorDetails.Donor.FirstNames,
				LastName:    donorDetails.Donor.LastName,
				DateOfBirth: donorDetails.Donor.DateOfBirth,
			}
		}

		if idActor == "voucher" {
			donorDetails.WantVoucher = form.Yes
		}

		if data.Voucher == "1" {
			donorDetails.Voucher = makeVoucher(voucherName)
			donorDetails.WantVoucher = form.Yes
		}

		attempts, err := strconv.Atoi(data.FailedVouchAttempts)
		if err != nil && data.FailedVouchAttempts != "" {
			return nil, nil, fmt.Errorf("invalid value for failedVouchAttempts: %s", err.Error())
		}

		donorDetails.FailedVouchAttempts = attempts
		donorDetails.IdentityUserData = userData
		if donorDetails.Tasks.ConfirmYourIdentity.IsNotStarted() {
			donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
		}
	}

	if data.Progress >= slices.Index(progressValues, "signTheLpa") {
		donorDetails.WantToApplyForLpa = true
		donorDetails.WantToSignLpa = true
		donorDetails.SignedAt = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
		donorDetails.WitnessedByCertificateProviderAt = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
		donorDetails.Tasks.SignTheLpa = task.StateCompleted
	}

	var certificateProviderUID actoruid.UID

	if data.Progress >= slices.Index(progressValues, "signedByCertificateProvider") {
		ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: random.String(16), LpaID: donorDetails.LpaID})

		certificateProvider, err := createCertificateProvider(ctx, shareCodeStore, certificateProviderStore, donorDetails.CertificateProvider.UID, donorDetails.SK, donorDetails.CertificateProvider.Email)
		if err != nil {
			return nil, nil, err
		}

		certificateProvider.ContactLanguagePreference = localize.En
		certificateProvider.SignedAt = donorDetails.SignedAt.AddDate(0, 0, 3)

		if err := certificateProviderStore.Put(ctx, certificateProvider); err != nil {
			return nil, nil, err
		}

		certificateProviderUID = certificateProvider.UID

		fns = append(fns, func(ctx context.Context, client *lpastore.Client, lpa *lpadata.Lpa) error {
			return client.SendCertificateProvider(ctx, certificateProvider, lpa)
		})
	}

	if data.Progress >= slices.Index(progressValues, "signedByAttorneys") {
		for isReplacement, list := range map[bool]donordata.Attorneys{false: donorDetails.Attorneys, true: donorDetails.ReplacementAttorneys} {
			for _, a := range list.Attorneys {
				ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: random.String(16), LpaID: donorDetails.LpaID})

				attorney, err := createAttorney(
					ctx,
					shareCodeStore,
					attorneyStore,
					a.UID,
					isReplacement,
					false,
					donorDetails.SK,
					a.Email,
				)
				if err != nil {
					return nil, nil, err
				}

				attorney.Phone = testMobile
				attorney.ContactLanguagePreference = localize.En
				attorney.Tasks.ConfirmYourDetails = task.StateCompleted
				attorney.Tasks.ReadTheLpa = task.StateCompleted
				attorney.Tasks.SignTheLpa = task.StateCompleted
				attorney.SignedAt = donorDetails.SignedAt.AddDate(0, 0, 10)

				if err := attorneyStore.Put(ctx, attorney); err != nil {
					return nil, nil, err
				}

				fns = append(fns, func(ctx context.Context, client *lpastore.Client, lpa *lpadata.Lpa) error {
					return client.SendAttorney(ctx, lpa, attorney)
				})
			}

			if list.TrustCorporation.Name != "" {
				ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: random.String(16), LpaID: donorDetails.LpaID})

				attorney, err := createAttorney(
					ctx,
					shareCodeStore,
					attorneyStore,
					list.TrustCorporation.UID,
					isReplacement,
					true,
					donorDetails.SK,
					list.TrustCorporation.Email,
				)
				if err != nil {
					return nil, nil, err
				}

				attorney.Phone = testMobile
				attorney.Tasks.ConfirmYourDetails = task.StateCompleted
				attorney.Tasks.ReadTheLpa = task.StateCompleted
				attorney.Tasks.SignTheLpa = task.StateCompleted
				attorney.WouldLikeSecondSignatory = form.No
				attorney.AuthorisedSignatories = [2]attorneydata.TrustCorporationSignatory{{
					FirstNames:        "A",
					LastName:          "Sign",
					ProfessionalTitle: "Assistant to the signer",
					SignedAt:          donorDetails.SignedAt.AddDate(0, 0, 15),
				}}

				if err := attorneyStore.Put(ctx, attorney); err != nil {
					return nil, nil, err
				}

				fns = append(fns, func(ctx context.Context, client *lpastore.Client, lpa *lpadata.Lpa) error {
					return client.SendAttorney(ctx, lpa, attorney)
				})
			}
		}
	}

	if data.Progress >= slices.Index(progressValues, "submitted") {
		donorDetails.SubmittedAt = time.Now()
	}

	if data.Progress >= slices.Index(progressValues, "statutoryWaitingPeriod") {
		fns = append(fns, func(ctx context.Context, client *lpastore.Client, _ *lpadata.Lpa) error {
			return client.SendStatutoryWaitingPeriod(ctx, donorDetails.LpaUID)
		})
		donorDetails.StatutoryWaitingPeriodAt = time.Now()
	}

	if data.Progress == slices.Index(progressValues, "registered") {
		fns = append(fns, func(ctx context.Context, client *lpastore.Client, _ *lpadata.Lpa) error {
			return client.SendRegister(ctx, donorDetails.LpaUID)
		})
	}

	if data.Progress == slices.Index(progressValues, "withdrawn") {
		fns = append(fns, func(ctx context.Context, client *lpastore.Client, _ *lpadata.Lpa) error {
			return client.SendChangeStatus(ctx, donorDetails.LpaUID, lpadata.StatusStatutoryWaitingPeriod, lpadata.StatusWithdrawn)
		})
		donorDetails.WithdrawnAt = time.Now()
	}

	if data.Progress == slices.Index(progressValues, "certificateProviderOptedOut") {
		fns = append(fns, func(ctx context.Context, client *lpastore.Client, _ *lpadata.Lpa) error {
			return client.SendCertificateProviderOptOut(ctx, donorDetails.LpaUID, certificateProviderUID)
		})
	}

	if data.Progress == slices.Index(progressValues, "doNotRegister") {
		fns = append(fns, func(ctx context.Context, client *lpastore.Client, _ *lpadata.Lpa) error {
			return client.SendChangeStatus(ctx, donorDetails.LpaUID, lpadata.StatusStatutoryWaitingPeriod, lpadata.StatusDoNotRegister)
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
		DonorEmail:                r.FormValue("donorEmail"),
		IdStatus:                  r.FormValue("idStatus"),
		Voucher:                   r.FormValue("voucher"),
		FailedVouchAttempts:       r.FormValue("failedVouchAttempts"),
	}
}
