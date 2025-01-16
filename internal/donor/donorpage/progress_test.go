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

	testCases := map[string]struct {
		donor                         *donordata.Provided
		setupCertificateProviderStore func(*mockCertificateProviderStore_GetAny_Call)
		setupVoucherStore             func(*mockVoucherStore_GetAny_Call)
		setupDonorStore               func(*testing.T) *mockDonorStore
		lpa                           *lpadata.Lpa
		infoNotifications             []progressNotification
		setupLocalizer                func(*testing.T) *mockLocalizer
		successNotifications          []progressNotification
	}{
		"none": {
			donor:                         &donordata.Provided{LpaUID: "lpa-uid"},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
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
			infoNotifications: []progressNotification{
				{
					Heading: "youHaveChosenToConfirmYourIdentityAtPostOffice",
					Body:    "whenYouHaveConfirmedAtPostOfficeReturnToTaskList",
				},
			},
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
		},

		// you've submitted your lpa to the opg
		"submitted": {
			donor: &donordata.Provided{
				LpaUID: "lpa-uid",
			},
			lpa: &lpadata.Lpa{
				LpaUID:    "lpa-uid",
				Submitted: true,
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []progressNotification{
				{
					Heading: "youveSubmittedYourLpaToOpg",
					Body:    "opgIsCheckingYourLpa",
				},
			},
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
		},
		"more evidence required": {
			donor: &donordata.Provided{
				Tasks:                  donordata.Tasks{PayForLpa: task.PaymentStateMoreEvidenceRequired},
				MoreEvidenceRequiredAt: testNow,
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []progressNotification{
				{
					Heading: "weNeedMoreEvidenceToMakeADecisionAboutYourLPAFee",
					Body:    "translated body",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("translated body")

				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")

				return l
			},
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
			infoNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "youDoNotNeedToTakeAnyAction",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"weHaveContactedVoucherToConfirmYourIdentity",
						map[string]any{"VoucherFullName": "a b"},
					).
					Return("translated heading")
				return l
			},
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
			infoNotifications: []progressNotification{
				{
					Heading: "youMustPayForYourLPA",
					Body:    "translated body",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"returnToTaskListToPayForLPAWeWillThenContactVoucher",
						map[string]any{"VoucherFullName": "a b"},
					).
					Return("translated body")
				return l
			},
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
			infoNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "translated body",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"voucherHasBeenUnableToConfirmYourIdentity",
						map[string]any{"VoucherFullName": "a b"},
					).
					Return("translated heading")

				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("translated body")

				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")

				return l
			},
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
			infoNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "translated body",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					T("thereIsAProblemWithYourLpa").
					Return("translated heading")
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("translated body")
				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")
				return l
			},
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
			infoNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "translated body",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					T("youHaveBeenUnableToConfirmYourIdentity").
					Return("translated heading")
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("translated body")
				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")
				return l
			},
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
			infoNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "translated body",
				},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"certificateProviderHasBeenUnableToConfirmIdentity",
						map[string]any{"CertificateProviderFullName": "c d"},
					).
					Return("translated heading")
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date"},
					).
					Return("translated body")
				l.EXPECT().
					FormatDate(testNow).
					Return("translated date")
				return l
			},
		},
		"voucher has vouched, lpa not signed": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
				Voucher: donordata.Voucher{FirstNames: "a", LastName: "b"},
			},
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
			successNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "returnToYourTaskListForInformationAboutWhatToDoNext",
				},
			},
			setupLocalizer: func(*testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"voucherHasConfirmedYourIdentity",
						map[string]any{"VoucherFullName": "c d"},
					).
					Return("translated heading")
				return l
			},
			setupVoucherStore: func(call *mockVoucherStore_GetAny_Call) {
				call.Return(&voucherdata.Provided{FirstNames: "c", LastName: "d", SignedAt: signedAt}, nil)
			},
			setupDonorStore: func(*testing.T) *mockDonorStore {
				s := newMockDonorStore(t)
				s.EXPECT().
					Put(context.Background(), &donordata.Provided{
						Tasks: donordata.Tasks{
							ConfirmYourIdentity: task.IdentityStateCompleted,
						},
						Voucher:                        donordata.Voucher{FirstNames: "a", LastName: "b"},
						HasViewedSuccessfulVouchBanner: true,
					}).
					Return(nil)
				return s
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
		},
		"voucher has vouched, lpa signed": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
					SignTheLpa:          task.StateCompleted,
				},
				Voucher: donordata.Voucher{FirstNames: "a", LastName: "b"},
			},
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
			successNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "youDoNotNeedToTakeAnyAction",
				},
			},
			setupLocalizer: func(*testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"voucherHasConfirmedYourIdentity",
						map[string]any{"VoucherFullName": "c d"},
					).
					Return("translated heading")
				return l
			},
			setupVoucherStore: func(call *mockVoucherStore_GetAny_Call) {
				call.Return(&voucherdata.Provided{FirstNames: "c", LastName: "d", SignedAt: signedAt}, nil)
			},
			setupDonorStore: func(*testing.T) *mockDonorStore {
				s := newMockDonorStore(t)
				s.EXPECT().
					Put(context.Background(), &donordata.Provided{
						Tasks: donordata.Tasks{
							ConfirmYourIdentity: task.IdentityStateCompleted,
							SignTheLpa:          task.StateCompleted,
						},
						Voucher:                        donordata.Voucher{FirstNames: "a", LastName: "b"},
						HasViewedSuccessfulVouchBanner: true,
					}).
					Return(nil)
				return s
			},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
		},
		"voucher has vouched, already seen notification": {
			donor: &donordata.Provided{
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
				Voucher:                        donordata.Voucher{FirstNames: "a", LastName: "b"},
				HasViewedSuccessfulVouchBanner: true,
			},
			lpa:                           &lpadata.Lpa{LpaUID: "lpa-uid"},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
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
			infoNotifications: []progressNotification{
				{Heading: "H", Body: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().T("weAreReviewingTheEvidenceYouSent").Return("H")
				l.EXPECT().T("ifYourEvidenceIsApprovedWillShowPaid").Return("B")
				return l
			},
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
			infoNotifications: []progressNotification{
				{Heading: "H", Body: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().T("yourLpaMustBeReviewedByCourtOfProtection").Return("H")
				l.EXPECT().T("opgIsCompletingChecksSoYouCanSubmitToCourtOfProtection").Return("B")
				return l
			},
		},
		"applying to court of protection and signed": {
			donor: &donordata.Provided{
				WitnessedByCertificateProviderAt: time.Now(),
				RegisteringWithCourtOfProtection: true,
			},
			lpa:                           &lpadata.Lpa{},
			setupCertificateProviderStore: certificateProviderStoreNotFound,
			infoNotifications: []progressNotification{
				{Heading: "H", Body: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().T("yourLpaMustBeReviewedByCourtOfProtection").Return("H")
				l.EXPECT().T("whenYouHavePaidOpgWillCheck").Return("B")
				return l
			},
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
			infoNotifications: []progressNotification{
				{Heading: "H", Body: "B"},
			},
			setupLocalizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().T("yourLpaMustBeReviewedByCourtOfProtection").Return("H")
				l.EXPECT().T("returnToYourTaskListToSignThenOpgWillCheck").Return("B")
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

			if tc.setupLocalizer != nil {
				testAppData.Localizer = tc.setupLocalizer(t)
			}

			voucherStore := newMockVoucherStore(t)
			if tc.setupVoucherStore != nil {
				tc.setupVoucherStore(voucherStore.EXPECT().GetAny(r.Context()))
			}

			donorStore := newMockDonorStore(t)
			if tc.setupDonorStore != nil {
				donorStore = tc.setupDonorStore(t)
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

			err := Progress(template.Execute, lpaStoreResolvingService, progressTracker, certificateProviderStore, voucherStore, donorStore)(testAppData, w, r, tc.donor)
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

	err := Progress(nil, lpaStoreResolvingService, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
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

	err := Progress(nil, lpaStoreResolvingService, nil, certificateProviderStore, nil, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetProgressWhenVoucherStoreErrors(t *testing.T) {
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

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := Progress(nil, lpaStoreResolvingService, progressTracker, certificateProviderStore, voucherStore, nil)(testAppData, w, r, &donordata.Provided{
		LpaUID:  "lpa-uid",
		Tasks:   donordata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
		Voucher: donordata.Voucher{FirstNames: "a"},
	})
	assert.Equal(t, expectedError, err)
}

func TestGetProgressWhenDonorStoreErrors(t *testing.T) {
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

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		GetAny(mock.Anything).
		Return(&voucherdata.Provided{SignedAt: time.Now()}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		Format(mock.Anything, mock.Anything).
		Return("")

	testAppData.Localizer = localizer

	err := Progress(nil, lpaStoreResolvingService, progressTracker, certificateProviderStore, voucherStore, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaUID:  "lpa-uid",
		Tasks:   donordata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
		Voucher: donordata.Voucher{FirstNames: "a"},
	})
	assert.Equal(t, expectedError, err)
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

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := Progress(template.Execute, lpaStoreResolvingService, progressTracker, certificateProviderStore, nil, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}
