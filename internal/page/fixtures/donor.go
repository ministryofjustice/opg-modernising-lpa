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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/document"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/reuse"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
)

type DynamoClient interface {
	OneByUID(ctx context.Context, uid string) (dynamo.Keys, error)
	Create(ctx context.Context, v interface{}) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
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
	"certificateProviderInvited",
	"certificateProviderAccessCodeUsed",
	"signedByCertificateProvider",
	"signedByAttorneys",
	"statutoryWaitingPeriod",
	// end states
	"registered",
	"withdrawn",
	"certificateProviderOptedOut",
	"doNotRegister",
}

type FixtureData struct {
	LpaType                    string
	Progress                   int
	Redirect                   string
	Donor                      string
	Attorneys                  string
	PeopleToNotify             string
	ReplacementAttorneys       string
	FeeType                    string
	PaymentTaskProgress        string
	WithVirus                  bool
	UseRealID                  bool
	CertificateProviderSub     string
	CertificateProviderEmail   string
	CertificateProviderMobile  string
	CertificateProviderChannel string
	DonorSub                   string
	DonorEmail                 string
	DonorMobile                string
	DonorFirstNames            string
	DonorLastName              string
	IdStatus                   string
	Voucher                    bool
	VouchAttempts              string
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
	accessCodeStore *accesscode.Store,
	voucherStore *voucher.Store,
	reuseStore *reuse.Store,
	notifyClient *notify.Client,
	appPublicURL string,
) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		data := setFixtureData(r)

		if data.DonorEmail == "" {
			data.DonorEmail = testEmail
		}

		if data.CertificateProviderChannel == "" {
			data.CertificateProviderChannel = "online"
		}

		if data.CertificateProviderEmail == "" && data.CertificateProviderChannel == "online" {
			data.CertificateProviderEmail = testEmail
		}

		if data.DonorSub == "" {
			data.DonorSub = random.AlphaNumeric(16)
		}

		if data.CertificateProviderSub == "" {
			data.CertificateProviderSub = random.AlphaNumeric(16)
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{
				App:        appData,
				Sub:        data.DonorSub,
				DonorEmail: data.DonorEmail,
				IdStatuses: []string{
					identity.StatusUnknown.String(),
					identity.StatusConfirmed.String(),
					identity.StatusFailed.String(),
					identity.StatusInsufficientEvidence.String(),
					identity.StatusExpired.String(),
					"post-office",
					"mismatch",
					"voucher-entered-code",
					"verified-not-vouched",
					"vouched",
					"vouch-failed",
				},
				CertificateProviderSub: data.CertificateProviderSub,
			})
		}

		encodedSub := encodeSub(data.DonorSub)
		donorSessionID := base64.StdEncoding.EncodeToString([]byte(mockGOLSubPrefix + encodedSub))

		if err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: mockGOLSubPrefix + encodedSub, Email: data.DonorEmail, HasLPAs: true}); err != nil {
			return err
		}

		donorDetails, err := donorStore.Create(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		donorCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, LpaID: donorDetails.LpaID})

		var fns []func(context.Context, *lpastore.Client, *lpadata.Lpa) error
		donorDetails, fns, err = updateLPAProgress(donorCtx, data, donorDetails, donorSessionID, r, certificateProviderStore, attorneyStore, documentStore, eventClient, accessCodeStore, voucherStore, reuseStore, notifyClient, appPublicURL, donorStore)
		if err != nil {
			return err
		}

		if data.Progress >= slices.Index(progressValues, "checkAndSendToYourCertificateProvider") {
			if err = donorDetails.UpdateCheckedHash(); err != nil {
				return fmt.Errorf("problem updating checkedHash: %w", err)
			}
		}

		if err := donorStore.Put(donorCtx, donorDetails); err != nil {
			return err
		}

		if !donorDetails.SignedAt.IsZero() && donorDetails.LpaUID != "" {
			if err := lpaStoreClient.SendLpa(donorCtx, donorDetails.LpaUID, lpastore.CreateLpaFromDonorProvided(donorDetails)); err != nil {
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

		http.Redirect(w, r, data.Redirect, http.StatusFound)
		return nil
	}
}

func updateLPAProgress(
	donorCtx context.Context,
	data FixtureData,
	donorDetails *donordata.Provided,
	donorSessionID string,
	r *http.Request,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	documentStore DocumentStore,
	eventClient *event.Client,
	accessCodeStore *accesscode.Store,
	voucherStore *voucher.Store,
	reuseStore *reuse.Store,
	notifyClient *notify.Client,
	appPublicURL string,
	donorStore DonorStore,
) (*donordata.Provided, []func(context.Context, *lpastore.Client, *lpadata.Lpa) error, error) {
	var fns []func(context.Context, *lpastore.Client, *lpadata.Lpa) error
	if data.Progress >= slices.Index(progressValues, "provideYourDetails") {
		if data.DonorFirstNames == "" {
			data.DonorFirstNames = "Sam"
		}

		if data.DonorLastName == "" {
			data.DonorLastName = "Smith"
		}

		donorDetails.Donor = makeDonor(data.DonorEmail, data.DonorFirstNames, data.DonorLastName)
		donorDetails.Donor.Mobile = data.DonorMobile

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

			donorDetails.LpaUID = waitForRealUID(15, donorStore, donorCtx)
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

	if attorneys := append(donorDetails.Attorneys.Attorneys, donorDetails.ReplacementAttorneys.Attorneys...); len(attorneys) > 0 {
		if err := reuseStore.PutAttorneys(donorCtx, attorneys); err != nil {
			return nil, nil, fmt.Errorf("reuse store put attorneys: %w", err)
		}
	}

	if trustCorporation := donorDetails.TrustCorporation(); trustCorporation.Name != "" {
		if err := reuseStore.PutTrustCorporation(donorCtx, trustCorporation); err != nil {
			return nil, nil, fmt.Errorf("reuse store put trust corporation: %w", err)
		}
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
		donorDetails.Restrictions = makeRestriction(donorDetails)
		donorDetails.Tasks.Restrictions = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "chooseYourCertificateProvider") {
		donorDetails.CertificateProvider = makeCertificateProvider()
		if data.CertificateProviderChannel == "paper" {
			donorDetails.CertificateProvider.CarryOutBy = lpadata.ChannelPaper
			donorDetails.CertificateProvider.Email = ""
		}

		if data.CertificateProviderEmail != "" && data.CertificateProviderChannel == "online" {
			donorDetails.CertificateProvider.Email = data.CertificateProviderEmail
		}

		if data.CertificateProviderMobile != "" {
			donorDetails.CertificateProvider.Mobile = data.CertificateProviderMobile
		}

		if err := reuseStore.PutCertificateProvider(donorCtx, donorDetails.CertificateProvider); err != nil {
			return nil, nil, fmt.Errorf("reuse store put certificate provider: %w", err)
		}

		donorDetails.Tasks.CertificateProvider = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "peopleToNotifyAboutYourLpa") {
		donorDetails.DoYouWantToNotifyPeople = form.Yes
		donorDetails.PeopleToNotify = []donordata.PersonToNotify{makePersonToNotify(peopleToNotifyNames[0]), makePersonToNotify(peopleToNotifyNames[1])}
		switch data.PeopleToNotify {
		case "without-address":
			donorDetails.PeopleToNotify[0].UID, _ = actoruid.Parse("f46f0ebf-794e-446f-9bcf-3aa72c929921")
			donorDetails.PeopleToNotify[0].Address = place.Address{}
		case "max":
			donorDetails.PeopleToNotify = append(donorDetails.PeopleToNotify, makePersonToNotify(peopleToNotifyNames[2]), makePersonToNotify(peopleToNotifyNames[3]), makePersonToNotify(peopleToNotifyNames[4]))
		}

		if err := reuseStore.PutPeopleToNotify(donorCtx, donorDetails.PeopleToNotify); err != nil {
			return nil, nil, fmt.Errorf("reuse store put people to notify: %w", err)
		}

		donorDetails.Tasks.PeopleToNotify = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "addCorrespondent") {
		donorDetails.AddCorrespondent = form.No
		donorDetails.Tasks.AddCorrespondent = task.StateCompleted
	}

	if data.Progress >= slices.Index(progressValues, "checkAndSendToYourCertificateProvider") {
		donorDetails.CheckedAt = time.Now()
		donorDetails.Tasks.CheckYourLpa = task.StateCompleted
		donorDetails.CertificateProviderInvitedAt = time.Now()
		donorDetails.CertificateProviderInvitedEmail = donorDetails.CertificateProvider.Email
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

			donorDetails.ReducedFeeDecisionAt = testNow
		} else {
			donorDetails.FeeType = pay.FullFee
		}

		donorDetails.PaymentDetails = append(donorDetails.PaymentDetails, donordata.Payment{
			PaymentReference: random.AlphaNumeric(12),
			PaymentID:        random.AlphaNumeric(12),
		})

		donorDetails.Tasks.PayForLpa = task.PaymentStateCompleted

		if data.PaymentTaskProgress != "" {
			taskState, err := task.ParsePaymentState(data.PaymentTaskProgress)
			if err != nil {
				return nil, nil, err
			}

			donorDetails.EvidenceDelivery = pay.Upload
			donorDetails.Tasks.PayForLpa = taskState

			if taskState.IsMoreEvidenceRequired() {
				donorDetails.MoreEvidenceRequiredAt = testNow
			}

			if taskState.IsApproved() {
				donorDetails.ReducedFeeDecisionAt = testNow
			}
		}
	}

	if data.Progress >= slices.Index(progressValues, "confirmYourIdentity") {
		var userData identity.UserData

		idActor, idStatus, ok := strings.Cut(data.IdStatus, ":")
		if !ok && data.IdStatus != "" {
			return nil, nil, errors.New("invalid value for idStatus - must be in format actor:status")
		}

		if idActor == "voucher" {
			donorDetails.WantVoucher = form.Yes
		}

		if data.Voucher {
			donorDetails.Voucher = makeVoucher(voucherName)
			donorDetails.WantVoucher = form.Yes

			if donorDetails.Tasks.PayForLpa.IsCompleted() {
				donorDetails.VoucherInvitedAt = time.Now()
				if donorDetails.Donor.Mobile != "" {
					donorDetails.VoucherCodeSentTo = donorDetails.Donor.Mobile
					donorDetails.VoucherCodeSentBySMS = true
				} else {
					donorDetails.VoucherCodeSentTo = donorDetails.Donor.Email
				}
			}
		}

		switch idStatus {
		case "failed":
			userData = identity.UserData{
				Status:    identity.StatusFailed,
				CheckedAt: time.Now(),
			}
		case "insufficient-evidence":
			userData = identity.UserData{
				Status:    identity.StatusInsufficientEvidence,
				CheckedAt: time.Now(),
			}
			donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStateInProgress
		case "expired":
			userData = identity.UserData{
				Status: identity.StatusExpired,
			}
		case "post-office":
			userData = identity.UserData{}
			donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStatePending
		case "voucher-entered-code", "verified-not-vouched", "vouched":
			userData = identity.UserData{
				Status:    identity.StatusInsufficientEvidence,
				CheckedAt: time.Now(),
			}

			donorDetails.Voucher = makeVoucher(voucherName)
			donorDetails.WantVoucher = form.Yes
			donorDetails.Voucher.Allowed = true
			donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStateInProgress

			ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: random.AlphaNumeric(16), LpaID: donorDetails.LpaID})

			voucherDetails, err := createVoucher(ctx, accessCodeStore, voucherStore, donorDetails)
			if err != nil {
				return nil, nil, fmt.Errorf("error creating voucher: %v", err)
			}

			voucherDetails.FirstNames = donorDetails.Voucher.FirstNames
			voucherDetails.LastName = donorDetails.Voucher.LastName

			if idStatus == "verified-not-vouched" || idStatus == "vouched" {
				donorDetails.DetailsVerifiedByVoucher = true
				donorDetails.VouchAttempts = 1

				voucherDetails.Tasks.ConfirmYourName = task.StateCompleted
				voucherDetails.Tasks.VerifyDonorDetails = task.StateCompleted
			}

			if idStatus == "vouched" {
				donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted

				voucherDetails.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
				voucherDetails.Tasks.SignTheDeclaration = task.StateCompleted
				voucherDetails.SignedAt = time.Now()
				voucherDetails.IdentityUserData = identity.UserData{
					Status:      identity.StatusConfirmed,
					FirstNames:  voucherDetails.FirstNames,
					LastName:    voucherDetails.LastName,
					DateOfBirth: donorDetails.Donor.DateOfBirth,
					CheckedAt:   time.Now(),
				}

				userData = identity.UserData{
					Status:         identity.StatusConfirmed,
					FirstNames:     donorDetails.Donor.FirstNames,
					LastName:       donorDetails.Donor.LastName,
					DateOfBirth:    donorDetails.Donor.DateOfBirth,
					CurrentAddress: donorDetails.Donor.Address,
					CheckedAt:      time.Now(),
				}
			}

			if err = voucherStore.Put(r.Context(), voucherDetails); err != nil {
				return nil, nil, fmt.Errorf("error persisting voucher: %v", err)
			}
		case "vouch-failed":
			donorDetails.WantVoucher = form.YesNoUnknown
			donorDetails.FailedVoucher = donorDetails.Voucher
			donorDetails.FailedVoucher.FailedAt = testNow
			donorDetails.Voucher = donordata.Voucher{}
			donorDetails.VoucherInvitedAt = time.Time{}

			userData = identity.UserData{
				Status:    identity.StatusInsufficientEvidence,
				CheckedAt: time.Now(),
			}

			donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStateInProgress
		case "mismatch":
			userData = identity.UserData{
				Status:      identity.StatusConfirmed,
				CheckedAt:   time.Now(),
				FirstNames:  donorDetails.Donor.FirstNames + " 1",
				LastName:    donorDetails.Donor.LastName + " 2",
				DateOfBirth: donorDetails.Donor.DateOfBirth,
			}
			donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStatePending
		default:
			userData = identity.UserData{
				Status:         identity.StatusConfirmed,
				CheckedAt:      time.Now(),
				FirstNames:     donorDetails.Donor.FirstNames,
				LastName:       donorDetails.Donor.LastName,
				DateOfBirth:    donorDetails.Donor.DateOfBirth,
				CurrentAddress: donorDetails.Donor.Address,
			}
		}

		if data.VouchAttempts != "" {
			attempts, err := strconv.Atoi(data.VouchAttempts)

			if err != nil {
				return nil, nil, fmt.Errorf("invalid value for vouchAttempts: %s", err.Error())
			}

			donorDetails.VouchAttempts = attempts
		}

		donorDetails.IdentityUserData = userData
		if donorDetails.Tasks.ConfirmYourIdentity.IsNotStarted() {
			donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
		}
	}

	if data.Progress >= slices.Index(progressValues, "signTheLpa") {
		donorDetails.SignedAt = time.Now()
		donorDetails.WitnessedByCertificateProviderAt = time.Now()

		if data.Donor == "signature-expired" {
			signedAt := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
			donorDetails.SignedAt = signedAt
			donorDetails.WitnessedByCertificateProviderAt = signedAt
		}

		donorDetails.WantToApplyForLpa = true
		donorDetails.WantToSignLpa = true
		donorDetails.Tasks.SignTheLpa = task.StateCompleted
	}

	certificateProviderEncodedSub := encodeSub(data.CertificateProviderSub)
	certificateProviderSessionID := base64.StdEncoding.EncodeToString([]byte(mockGOLSubPrefix + certificateProviderEncodedSub))
	certificateProviderCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: certificateProviderSessionID, LpaID: donorDetails.LpaID})

	if data.Progress >= slices.Index(progressValues, "certificateProviderInvited") {
		plainCode, hashedCode := accesscodedata.Generate()
		accessCodeData := accesscodedata.Link{
			PK:          dynamo.AccessKey(dynamo.CertificateProviderAccessKey(hashedCode.String())),
			SK:          dynamo.ShareSortKey(dynamo.MetadataKey(hashedCode.String())),
			ActorUID:    donorDetails.CertificateProvider.UID,
			LpaOwnerKey: donorDetails.SK,
			LpaUID:      donorDetails.LpaUID,
			LpaKey:      donorDetails.PK,
		}

		err := accessCodeStore.Put(certificateProviderCtx, actor.TypeCertificateProvider, hashedCode, accessCodeData)
		if err != nil {
			return nil, nil, err
		}

		if err := notifyClient.SendEmail(certificateProviderCtx, notify.ToCustomEmail(localize.En, data.CertificateProviderEmail), notify.CertificateProviderInviteEmail{
			DonorFullName:                donorDetails.Donor.FullName(),
			LpaType:                      "Property and affairs",
			CertificateProviderFullName:  donorDetails.CertificateProvider.FullName(),
			DonorFirstNames:              donorDetails.Donor.FirstNames,
			DonorFirstNamesPossessive:    donorDetails.Donor.FirstNames + "’s",
			WhatLpaCovers:                "money, finances and any property they might own",
			CertificateProviderStartURL:  appPublicURL + page.PathCertificateProviderEnterAccessCode.Format(),
			AccessCode:                   plainCode.Plain(),
			CertificateProviderOptOutURL: appPublicURL + page.PathCertificateProviderEnterAccessCodeOptOut.Format(),
		}); err != nil {
			return nil, nil, err
		}
	}

	var certificateProviderUID actoruid.UID
	var certificateProvider *certificateproviderdata.Provided

	if data.Progress >= slices.Index(progressValues, "certificateProviderAccessCodeUsed") {
		var err error
		certificateProvider, err = createCertificateProvider(certificateProviderCtx, accessCodeStore, certificateProviderStore, donorDetails)
		if err != nil {
			return nil, nil, err
		}

		certificateProvider.ContactLanguagePreference = localize.En
		certificateProvider.SignedAt = donorDetails.SignedAt.AddDate(0, 0, 3)

		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return nil, nil, err
		}

		certificateProviderUID = certificateProvider.UID
	}

	if data.Progress >= slices.Index(progressValues, "signedByCertificateProvider") {
		if donorDetails.CertificateProvider.CarryOutBy.IsOnline() {
			fns = append(fns, func(ctx context.Context, client *lpastore.Client, lpa *lpadata.Lpa) error {
				return client.SendCertificateProvider(ctx, certificateProvider, lpa)
			})
		} else {
			fns = append(fns, func(ctx context.Context, client *lpastore.Client, lpa *lpadata.Lpa) error {
				return client.SendPaperCertificateProviderSign(ctx, lpa.LpaUID, donorDetails.CertificateProvider)
			})
		}
	}

	if data.Progress == slices.Index(progressValues, "certificateProviderOptedOut") {
		fns = append(fns, func(ctx context.Context, client *lpastore.Client, _ *lpadata.Lpa) error {
			return client.SendCertificateProviderOptOut(ctx, donorDetails.LpaUID, certificateProviderUID)
		})

		return donorDetails, fns, nil
	}

	if data.Progress >= slices.Index(progressValues, "signedByAttorneys") {
		for isReplacement, list := range map[bool]donordata.Attorneys{false: donorDetails.Attorneys, true: donorDetails.ReplacementAttorneys} {
			for _, a := range list.Attorneys {
				ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: random.AlphaNumeric(16), LpaID: donorDetails.LpaID})

				attorney, err := createAttorney(
					ctx,
					accessCodeStore,
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
				ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: random.AlphaNumeric(16), LpaID: donorDetails.LpaID})

				attorney, err := createAttorney(
					ctx,
					accessCodeStore,
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
			return client.SendDonorWithdrawLPA(ctx, donorDetails.LpaUID)
		})
		donorDetails.WithdrawnAt = testNow
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
		donorDetails.DoNotRegisterAt = time.Now()
	}

	return donorDetails, fns, nil
}

func setFixtureData(r *http.Request) FixtureData {
	r.ParseForm()
	options := r.Form["options"]

	paymentTaskProgress := r.FormValue("paymentTaskProgress")
	if slices.Contains(options, "paymentTaskInProgress") {
		paymentTaskProgress = "InProgress"
	}

	return FixtureData{
		LpaType:                    r.FormValue("lpa-type"),
		Progress:                   slices.Index(progressValues, r.FormValue("progress")),
		Redirect:                   r.FormValue("redirect"),
		Donor:                      r.FormValue("donor"),
		Attorneys:                  r.FormValue("attorneys"),
		PeopleToNotify:             r.FormValue("peopleToNotify"),
		ReplacementAttorneys:       r.FormValue("replacementAttorneys"),
		FeeType:                    r.FormValue("feeType"),
		PaymentTaskProgress:        paymentTaskProgress,
		WithVirus:                  r.FormValue("withVirus") == "1",
		UseRealID:                  slices.Contains(options, "uid"),
		CertificateProviderSub:     r.FormValue("certificateProviderSub"),
		CertificateProviderEmail:   r.FormValue("certificateProviderEmail"),
		CertificateProviderMobile:  r.FormValue("certificateProviderMobile"),
		CertificateProviderChannel: r.FormValue("certificateProviderChannel"),
		DonorSub:                   r.FormValue("donorSub"),
		DonorEmail:                 r.FormValue("donorEmail"),
		DonorMobile:                r.FormValue("donorMobile"),
		DonorFirstNames:            r.FormValue("donorFirstNames"),
		DonorLastName:              r.FormValue("donorLastName"),
		IdStatus:                   r.FormValue("idStatus"),
		Voucher:                    slices.Contains(options, "voucher"),
		VouchAttempts:              r.FormValue("vouchAttempts"),
	}
}
