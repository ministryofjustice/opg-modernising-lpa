package donorpage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProgress(t *testing.T) {
	signedAt := time.Now()

	certificateProviderStoreNotFound := func(call *mockCertificateProviderStore_GetAny_Call) {
		call.Return(nil, dynamo.NotFoundError{})
	}

	voucherStoreNotFound := func(call *mockVoucherStore_GetAny_Call) {
		call.Return(nil, dynamo.NotFoundError{})
	}

	donorStoreNoUpdate := func(_ *testing.T, s *mockDonorStore) {
		s.EXPECT().
			Put(mock.Anything, mock.Anything).
			Return(nil)
	}

	testCases := map[string]struct {
		donor                         *donordata.Provided
		setupCertificateProviderStore func(*mockCertificateProviderStore_GetAny_Call)
		setupVoucherStore             func(*mockVoucherStore_GetAny_Call)
		setupDonorStore               func(*testing.T, *mockDonorStore)
		lpa                           *lpadata.Lpa
		infoNotifications             []page.Notification
		successNotifications          []page.Notification
		setupLocalizer                func(*testing.T) *mockLocalizer
	}{
		"none": {
			donor:                         &donordata.Provided{LpaUID: "lpa-uid"},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			setupDonorStore:               donorStoreNoUpdate,
		},

		// you have chosen to confirm your identity at a post office
		"going to the post office": {
			donor: &donordata.Provided{
				LpaUID: "lpa-uid",
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStatePending,
				},
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "youHaveChosenToConfirmYourIdentityAtPostOffice",
					BodyHTML: "whenYouHaveConfirmedAtPostOfficeReturnToTaskList",
				},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"confirmed identity": {
			donor: &donordata.Provided{
				LpaUID: "lpa-uid",
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			setupDonorStore:               donorStoreNoUpdate,
		},

		// you've submitted your lpa to the opg
		"submitted": {
			donor: &donordata.Provided{
				LpaUID:                           "lpa-uid",
				WitnessedByCertificateProviderAt: signedAt,
			},
			lpa: &lpadata.Lpa{
				LpaUID:    "lpa-uid",
				Submitted: true,
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "youveSubmittedYourLpaToOpg",
					BodyHTML: "opgIsCheckingYourLpa",
				},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"submitted and certificate provider started": {
			donor: &donordata.Provided{
				LpaUID: "lpa-uid",
			},
			lpa: &lpadata.Lpa{
				LpaUID:    "lpa-uid",
				Submitted: true,
			},
			setupCertificateProviderStore: func(call *mockCertificateProviderStore_GetAny_Call) {
				call.Return(&certificateproviderdata.Provided{}, nil)
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"submitted and certificate provider finished": {
			donor: &donordata.Provided{
				LpaUID: "lpa-uid",
			},
			lpa: &lpadata.Lpa{
				LpaUID:    "lpa-uid",
				Submitted: true,
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
			setupCertificateProviderStore: func(call *mockCertificateProviderStore_GetAny_Call) {
				call.Return(&certificateproviderdata.Provided{}, nil)
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"more evidence required": {
			donor: &donordata.Provided{
				Tasks:                  donordata.Tasks{PayForLpa: task.PaymentStateMoreEvidenceRequired},
				MoreEvidenceRequiredAt: testNow,
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "weNeedMoreEvidenceToMakeADecisionAboutYourLPAFee",
					BodyHTML: "B",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("B")

				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")

				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"voucher has been contacted": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateInProgress,
				},
				VoucherInvitedAt: testNow,
				Voucher:          donordata.Voucher{FirstNames: "a", LastName: "b"},
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "H",
					BodyHTML: "youDoNotNeedToTakeAnyAction",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"weHaveContactedVoucherToConfirmYourIdentity",
						map[string]any{"VoucherFullName": "a b"},
					).
					Return("H")
				return l
			},
			setupVoucherStore: voucherStoreNotFound,
			setupDonorStore:   donorStoreNoUpdate,
		},
		"voucher has been chosen but not contacted": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateInProgress,
					PayForLpa:           task.PaymentStateInProgress,
				},
				Voucher: donordata.Voucher{FirstNames: "a", LastName: "b"},
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "youMustPayForYourLPA",
					BodyHTML: "B",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"returnToTaskListToPayForLPAWeWillThenContactVoucher",
						map[string]any{"VoucherFullName": "a b"},
					).
					Return("B")
				return l
			},
			setupVoucherStore: voucherStoreNotFound,
			setupDonorStore:   donorStoreNoUpdate,
		},
		"voucher was unsuccessful": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{ConfirmYourIdentity: task.IdentityStateInProgress},
				FailedVoucher: donordata.Voucher{
					FirstNames: "a",
					LastName:   "b",
					FailedAt:   testNow,
				},
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "H",
					BodyHTML: "B",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"voucherHasBeenUnableToConfirmYourIdentity",
						map[string]any{"VoucherFullName": "a b"},
					).
					Return("H")

				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("B")

				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")

				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"do not register": {
			donor: &donordata.Provided{
				LpaUID:          "lpa-uid",
				DoNotRegisterAt: testNow,
			},
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Status: lpadata.StatusDoNotRegister,
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "thereIsAProblemWithYourLpa",
					BodyHTML: "B",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("B")
				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")
				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"donor identity check failed": {
			donor: &donordata.Provided{
				LpaUID: "lpa-uid",
				IdentityUserData: identity.UserData{
					Status:    identity.StatusFailed,
					CheckedAt: testNow,
				},
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "youHaveBeenUnableToConfirmYourIdentity",
					BodyHTML: "B",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("B")
				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")
				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"certificate provider identity check failed": {
			donor: &donordata.Provided{LpaUID: "lpa-uid"},
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
				},
			},
			setupCertificateProviderStore: func(call *mockCertificateProviderStore_GetAny_Call) {
				call.Return(&certificateproviderdata.Provided{
					IdentityUserData: identity.UserData{
						Status:    identity.StatusFailed,
						CheckedAt: testNow,
					},
				}, nil)
			},
			infoNotifications: []page.Notification{
				{
					Heading:  "H",
					BodyHTML: "B",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"certificateProviderHasBeenUnableToConfirmIdentity",
						map[string]any{"CertificateProviderFullName": "c d"},
					).
					Return("H")
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("B")
				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")
				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"voucher has vouched, lpa not signed": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
				Voucher: donordata.Voucher{FirstNames: "a", LastName: "b"},
			},
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
			successNotifications: []page.Notification{
				{
					Heading:  "H",
					BodyHTML: "returnToYourTaskListForInformationAboutWhatToDoNext",
				},
			},
			setupLocalizer: func(*testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"voucherHasConfirmedYourIdentity",
						map[string]any{"VoucherFullName": "c d"},
					).
					Return("H")
				return l
			},
			setupVoucherStore: func(call *mockVoucherStore_GetAny_Call) {
				call.Return(&voucherdata.Provided{FirstNames: "c", LastName: "d", SignedAt: signedAt}, nil)
			},
			setupDonorStore: func(_ *testing.T, s *mockDonorStore) {
				s.EXPECT().
					Put(context.Background(), &donordata.Provided{
						Tasks: donordata.Tasks{
							ConfirmYourIdentity: task.IdentityStateCompleted,
						},
						Voucher:                      donordata.Voucher{FirstNames: "a", LastName: "b"},
						HasSeenSuccessfulVouchBanner: true,
					}).
					Return(nil)
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
		},
		"voucher has vouched, lpa signed": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
					SignTheLpa:          task.StateCompleted,
				},
				Voucher:                          donordata.Voucher{FirstNames: "a", LastName: "b"},
				WitnessedByCertificateProviderAt: signedAt,
			},
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
			successNotifications: []page.Notification{
				{
					Heading:  "H",
					BodyHTML: "youDoNotNeedToTakeAnyAction",
				},
			},
			setupLocalizer: func(*testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"voucherHasConfirmedYourIdentity",
						map[string]any{"VoucherFullName": "c d"},
					).
					Return("H")
				return l
			},
			setupVoucherStore: func(call *mockVoucherStore_GetAny_Call) {
				call.Return(&voucherdata.Provided{FirstNames: "c", LastName: "d", SignedAt: signedAt}, nil)
			},
			setupDonorStore: func(_ *testing.T, s *mockDonorStore) {
				s.EXPECT().
					Put(context.Background(), &donordata.Provided{
						Tasks: donordata.Tasks{
							ConfirmYourIdentity: task.IdentityStateCompleted,
							SignTheLpa:          task.StateCompleted,
						},
						Voucher:                          donordata.Voucher{FirstNames: "a", LastName: "b"},
						HasSeenSuccessfulVouchBanner:     true,
						WitnessedByCertificateProviderAt: signedAt,
					}).
					Return(nil)
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
		},
		"voucher has vouched, already seen notification": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
				Voucher:                      donordata.Voucher{FirstNames: "a", LastName: "b"},
				HasSeenSuccessfulVouchBanner: true,
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			setupVoucherStore: func(call *mockVoucherStore_GetAny_Call) {
				call.Return(&voucherdata.Provided{FirstNames: "c", LastName: "d", SignedAt: signedAt}, nil)
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"reduced fee approved payment task complete": {
			donor: &donordata.Provided{
				LpaUID:               "lpa-uid",
				Tasks:                donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
				ReducedFeeDecisionAt: testNow,
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			successNotifications: []page.Notification{
				{
					Heading:  "weHaveApprovedYourLPAFeeRequest",
					BodyHTML: "yourLPAIsNowPaid",
				},
			},
			setupDonorStore: func(_ *testing.T, s *mockDonorStore) {
				s.EXPECT().
					Put(mock.Anything, &donordata.Provided{
						LpaUID:                                "lpa-uid",
						Tasks:                                 donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
						ReducedFeeDecisionAt:                  testNow,
						HasSeenReducedFeeApprovalNotification: true,
					}).
					Return(nil)
			},
		},
		"reduced fee approved payment task complete - has seen notification": {
			donor: &donordata.Provided{
				LpaUID:                                "lpa-uid",
				Tasks:                                 donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
				ReducedFeeDecisionAt:                  testNow,
				HasSeenReducedFeeApprovalNotification: true,
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			setupDonorStore:               donorStoreNoUpdate,
		},
		"applied for reduced fee": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStatePending,
				},
				FeeType:        pay.HalfFee,
				PaymentDetails: []donordata.Payment{{Amount: 4100}},
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "weAreReviewingTheEvidenceYouSent", BodyHTML: "ifYourEvidenceIsApprovedWillShowPaid"},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"denied reduced fee": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStateDenied,
				},
				FeeType:              pay.FullFee,
				ReducedFeeDecisionAt: testNow,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "thereIsFeeToPay", BodyHTML: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().Format("weContactedYouToExplainWhyNotEligible", map[string]any{"ContactedDate": "translated date"}).Return("B")
				l.EXPECT().FormatDate(testNow).Return("translated date")
				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"approved other reduced fee": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStateApproved,
				},
				FeeType:              pay.HalfFee,
				ReducedFeeDecisionAt: testNow,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "thereIsFeeToPay", BodyHTML: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().Format("weContactedYouToExplainWhyNotEligibleNowPay", map[string]any{"ContactedDate": "translated date"}).Return("B")
				l.EXPECT().FormatDate(testNow).Return("translated date")
				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"applying to court of protection and signed and paid": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStateCompleted,
				},
				WitnessedByCertificateProviderAt: time.Now(),
				RegisteringWithCourtOfProtection: true,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "yourLpaMustBeReviewedByCourtOfProtection", BodyHTML: "opgIsCompletingChecksSoYouCanSubmitToCourtOfProtection"},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"applying to court of protection and signed": {
			donor: &donordata.Provided{
				WitnessedByCertificateProviderAt: time.Now(),
				RegisteringWithCourtOfProtection: true,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "yourLpaMustBeReviewedByCourtOfProtection", BodyHTML: "whenYouHavePaidOpgWillCheck"},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"applying to court of protection and paid": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStateCompleted,
				},
				RegisteringWithCourtOfProtection: true,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "yourLpaMustBeReviewedByCourtOfProtection", BodyHTML: "returnToYourTaskListToSignThenOpgWillCheck"},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"withdrawn": {
			donor: &donordata.Provided{
				LpaUID:      "lpa-uid",
				WithdrawnAt: testNow,
			},
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Status: lpadata.StatusWithdrawn,
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{
					Heading:  "lpaRevoked",
					BodyHTML: "translated body",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"weContactedYouOnAboutLPARevokedOPGWillNotRegister",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("translated body")
				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")
				return l
			},
		},
		"identity not confirmed and LPA signed - exceeded identity deadline date": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					SignTheLpa: task.StateCompleted,
				},
				WitnessedByCertificateProviderAt: signedAt.AddDate(0, -6, -1),
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "yourLPACannotBeRegisteredByOPG", BodyHTML: "youDidNotConfirmYourIdentityWithinSixMonthsOfSigning"},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"identity expired and LPA not signed": {
			donor: &donordata.Provided{
				Tasks:            donordata.Tasks{},
				IdentityUserData: identity.UserData{Status: identity.StatusExpired},
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "youMustConfirmYourIdentityAgain", BodyHTML: "youDidNotSignYourLPAWithinSixMonthsOfConfirmingYourIdentity"},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"statutory waiting period": {
			donor: &donordata.Provided{},
			lpa: &lpadata.Lpa{
				Status: lpadata.StatusStatutoryWaitingPeriod,
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "yourLpaIsAwaitingRegistration", BodyHTML: "theOpgWillRegisterYourLpaAtEndOfWaitingPeriod"},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"identity mismatch pending": {
			donor: &donordata.Provided{
				ContinueWithMismatchedIdentity: true,
				IdentityUserData:               identity.UserData{Status: identity.StatusConfirmed},
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStatePending,
				},
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "confirmationOfIdentityPending", BodyHTML: "youDoNotNeedToTakeAnyAction"},
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"priority correspondence sent": {
			donor: &donordata.Provided{
				PriorityCorrespondenceSentAt: testNow,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []page.Notification{
				{Heading: "thereIsAProblemWithYourLpa", BodyHTML: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().Format("weContactedYouOnWithGuidanceAboutWhatToDoNext", map[string]any{"ContactedDate": "translated date"}).Return("B")
				l.EXPECT().FormatDate(testNow).Return("translated date")
				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"priority correspondence sent registered": {
			donor: &donordata.Provided{
				PriorityCorrespondenceSentAt: testNow,
			},
			lpa: &lpadata.Lpa{
				Status: lpadata.StatusRegistered,
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			setupDonorStore:               donorStoreNoUpdate,
		},
		"identity mismatch resolved": {
			donor: &donordata.Provided{
				ContinueWithMismatchedIdentity: true,
				IdentityUserData:               identity.UserData{Status: identity.StatusConfirmed},
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			successNotifications: []page.Notification{
				{Heading: "yourIdentityHadBeenConfirmed", BodyHTML: "youDoNotNeedToTakeAnyAction"},
			},
			setupDonorStore: func(_ *testing.T, s *mockDonorStore) {
				s.EXPECT().
					Put(mock.Anything, &donordata.Provided{
						ContinueWithMismatchedIdentity: true,
						IdentityUserData:               identity.UserData{Status: identity.StatusConfirmed},
						Tasks: donordata.Tasks{
							ConfirmYourIdentity: task.IdentityStateCompleted,
						},
						HasSeenIdentityMismatchResolvedNotification: true,
					}).
					Return(nil)
			},
		},
		"identity mismatch resolved when already seen": {
			donor: &donordata.Provided{
				ContinueWithMismatchedIdentity: true,
				IdentityUserData:               identity.UserData{Status: identity.StatusConfirmed},
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
				HasSeenIdentityMismatchResolvedNotification: true,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			setupDonorStore:               donorStoreNoUpdate,
		},
		"identity mismatch material change confirmed": {
			donor: &donordata.Provided{
				ContinueWithMismatchedIdentity: true,
				IdentityUserData:               identity.UserData{Status: identity.StatusConfirmed},
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateProblem,
				},
				MaterialChangeConfirmedAt: testNow,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			setupDonorStore:               donorStoreNoUpdate,
			infoNotifications: []page.Notification{
				{Heading: "yourLPACannotBeRegisteredByOPG", BodyHTML: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().Format("weContactedYouOnWithGuidanceAboutWhatToDoNext", map[string]any{"ContactedDate": "translated date"}).Return("B")
				l.EXPECT().FormatDate(testNow).Return("translated date")
				return l
			},
		},
		"certificate provider identity mismatch pending": {
			donor: &donordata.Provided{},
			lpa: &lpadata.Lpa{
				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
			},
			setupCertificateProviderStore: func(call *mockCertificateProviderStore_GetAny_Call) {
				call.Return(&certificateproviderdata.Provided{
					IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
					Tasks: certificateproviderdata.Tasks{
						ConfirmYourIdentity: task.IdentityStatePending,
					},
				}, nil)
			},
			infoNotifications: []page.Notification{
				{Heading: "heading", BodyHTML: "wellContactYouIfYouNeedToTakeAnyAction"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().Format(
					"certificateProviderConfirmationOfIdentityPending",
					map[string]any{"CertificateProviderFullName": "A B"},
				).Return("heading")
				return l
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"certificate provider identity mismatch resolved": {
			donor: &donordata.Provided{},
			lpa: &lpadata.Lpa{
				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
			},
			setupCertificateProviderStore: func(call *mockCertificateProviderStore_GetAny_Call) {
				call.Return(&certificateproviderdata.Provided{
					ImmaterialChangeConfirmedAt: testNow,
					IdentityUserData:            identity.UserData{Status: identity.StatusConfirmed},
					Tasks: certificateproviderdata.Tasks{
						ConfirmYourIdentity: task.IdentityStateCompleted,
					},
				}, nil)
			},
			successNotifications: []page.Notification{
				{Heading: "heading", BodyHTML: "youDoNotNeedToTakeAnyAction"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().Format(
					"certificateProviderIdentityConfirmed",
					map[string]any{"CertificateProviderFullName": "A B"},
				).Return("heading")
				return l
			},
			setupDonorStore: func(_ *testing.T, s *mockDonorStore) {
				s.EXPECT().
					Put(mock.Anything, &donordata.Provided{
						HasSeenCertificateProviderIdentityMismatchResolvedNotification: true,
					}).
					Return(nil)
			},
		},
		"certificate provider identity mismatch resolved when already seen": {
			donor: &donordata.Provided{
				HasSeenCertificateProviderIdentityMismatchResolvedNotification: true,
			},
			lpa: &lpadata.Lpa{},
			setupCertificateProviderStore: func(call *mockCertificateProviderStore_GetAny_Call) {
				call.Return(&certificateproviderdata.Provided{
					ImmaterialChangeConfirmedAt: testNow,
					IdentityUserData:            identity.UserData{Status: identity.StatusConfirmed},
					Tasks: certificateproviderdata.Tasks{
						ConfirmYourIdentity: task.IdentityStateCompleted,
					},
				}, nil)
			},
			setupDonorStore: donorStoreNoUpdate,
		},
		"certificate provider identity mismatch material change confirmed": {
			donor: &donordata.Provided{},
			lpa:   &lpadata.Lpa{},
			setupCertificateProviderStore: func(call *mockCertificateProviderStore_GetAny_Call) {
				call.Return(&certificateproviderdata.Provided{
					MaterialChangeConfirmedAt: testNow,
					IdentityUserData:          identity.UserData{Status: identity.StatusConfirmed},
					Tasks: certificateproviderdata.Tasks{
						ConfirmYourIdentity: task.IdentityStateProblem,
					},
				}, nil)
			},
			setupDonorStore: donorStoreNoUpdate,
			infoNotifications: []page.Notification{
				{Heading: "yourLPACannotBeRegisteredByOPG", BodyHTML: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().Format("weContactedYouOnWithGuidanceAboutWhatToDoNext", map[string]any{"ContactedDate": "translated date"}).Return("B")
				l.EXPECT().FormatDate(testNow).Return("translated date")
				return l
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.lpa, nil)

			progressTracker := newMockProgressTracker(t)
			progressTracker.EXPECT().
				Progress(tc.lpa).
				Return(task.Progress{DonorSigned: task.ProgressTask{Done: true}})

			certificateProviderStore := newMockCertificateProviderStore(t)
			tc.setupCertificateProviderStore(certificateProviderStore.EXPECT().
				GetAny(r.Context()))

			donorStore := newMockDonorStore(t)
			if tc.setupDonorStore != nil {
				tc.setupDonorStore(t, donorStore)
			}

			if tc.setupLocalizer != nil {
				testAppData.Localizer = tc.setupLocalizer(t)
			}

			voucherStore := newMockVoucherStore(t)
			if tc.setupVoucherStore != nil {
				tc.setupVoucherStore(voucherStore.EXPECT().GetAny(r.Context()))
			}

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &progressData{
					App:                  testAppData,
					Donor:                tc.donor,
					Progress:             task.Progress{DonorSigned: task.ProgressTask{Done: true}},
					InfoNotifications:    tc.infoNotifications,
					SuccessNotifications: tc.successNotifications,
				}).
				Return(nil)

			err := Progress(template.Execute, lpaStoreResolvingService, progressTracker, certificateProviderStore, voucherStore, donorStore, time.Now)(testAppData, w, r, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetProgressWhenLpaStoreClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	err := Progress(nil, lpaStoreResolvingService, nil, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.ErrorIs(t, err, expectedError)
}

func TestGetProgressWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{Submitted: true}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := Progress(nil, lpaStoreResolvingService, nil, certificateProviderStore, nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.ErrorIs(t, err, expectedError)
}

func TestGetProgressWhenVoucherStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := Progress(nil, lpaStoreResolvingService, nil, certificateProviderStore, voucherStore, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaUID:  "lpa-uid",
		Tasks:   donordata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
		Voucher: donordata.Voucher{FirstNames: "a"},
	})
	assert.ErrorIs(t, err, expectedError)
}

func TestGetProgressOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything).
		Return(task.Progress{})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := Progress(template.Execute, lpaStoreResolvingService, progressTracker, certificateProviderStore, nil, donorStore, testNowFn)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetProgressOnDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything).
		Return(task.Progress{})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := Progress(nil, lpaStoreResolvingService, progressTracker, certificateProviderStore, nil, donorStore, time.Now)(testAppData, w, r, &donordata.Provided{
		LpaUID:               "lpa-uid",
		Tasks:                donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
		ReducedFeeDecisionAt: time.Now(),
	})

	assert.ErrorContains(t, err, "failed to update donor: err")
}
