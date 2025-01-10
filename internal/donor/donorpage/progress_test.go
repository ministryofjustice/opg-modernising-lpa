package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProgress(t *testing.T) {
	signedAt := time.Now()

	testCases := map[string]struct {
		donor                         *donordata.Provided
		setupCertificateProviderStore func(*mockCertificateProviderStore_GetAny_Call)
		lpa                           *lpadata.Lpa
		infoNotifications             []progressNotification
		setupLocalizer                func() *mockLocalizer
	}{
		"none": {
			donor: &donordata.Provided{LpaUID: "lpa-uid"},
			lpa:   &lpadata.Lpa{LpaUID: "lpa-uid"},
		},

		// you have chosen to confirm your identity at a post office
		"going to the post office": {
			donor: &donordata.Provided{
				LpaUID: "lpa-uid",
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStatePending,
				},
			},
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
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
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
		},

		// you've submitted your lpa to the opg
		"submitted": {
			donor: &donordata.Provided{
				LpaUID:      "lpa-uid",
				SubmittedAt: time.Now(),
			},
			lpa: &lpadata.Lpa{
				LpaUID:    "lpa-uid",
				Submitted: true,
			},
			setupCertificateProviderStore: func(call *mockCertificateProviderStore_GetAny_Call) {
				call.Return(nil, dynamo.NotFoundError{})
			},
			infoNotifications: []progressNotification{
				{
					Heading: "youveSubmittedYourLpaToOpg",
					Body:    "opgIsCheckingYourLpa",
				},
			},
		},
		"submitted and certificate provider started": {
			donor: &donordata.Provided{
				LpaUID:      "lpa-uid",
				SubmittedAt: time.Now(),
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
				LpaUID:      "lpa-uid",
				SubmittedAt: time.Now(),
			},
			lpa: &lpadata.Lpa{
				LpaUID:    "lpa-uid",
				Submitted: true,
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			},
		},
		"more evidence required": {
			donor: &donordata.Provided{
				Tasks:                  donordata.Tasks{PayForLpa: task.PaymentStateMoreEvidenceRequired},
				MoreEvidenceRequiredAt: testNow,
			},
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
			infoNotifications: []progressNotification{
				{
					Heading: "weNeedMoreEvidenceToMakeADecisionAboutYourLPAFee",
					Body:    "translated body",
				},
			},
			setupLocalizer: func() *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": "translated date time"},
					).
					Return("translated body")

				l.EXPECT().
					FormatDateTime(testNow).
					Return("translated date time")

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
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
			infoNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "youDoNotNeedToTakeAnyAction",
				},
			},
			setupLocalizer: func() *mockLocalizer {
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
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
			infoNotifications: []progressNotification{
				{
					Heading: "youMustPayForYourLPA",
					Body:    "translated body",
				},
			},
			setupLocalizer: func() *mockLocalizer {
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
			lpa: &lpadata.Lpa{LpaUID: "lpa-uid"},
			infoNotifications: []progressNotification{
				{
					Heading: "translated heading",
					Body:    "translated body",
				},
			},
			setupLocalizer: func() *mockLocalizer {
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
						map[string]any{"ContactedDate": "translated date time"},
					).
					Return("translated body")

				l.EXPECT().
					FormatDateTime(testNow).
					Return("translated date time")

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
			if tc.setupCertificateProviderStore != nil {
				tc.setupCertificateProviderStore(certificateProviderStore.EXPECT().GetAny(r.Context()))
			}

			if tc.setupLocalizer != nil {
				testAppData.Localizer = tc.setupLocalizer()
			}

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &progressData{
					App:               testAppData,
					Donor:             tc.donor,
					Progress:          task.Progress{DonorSigned: task.ProgressTask{Done: true}},
					InfoNotifications: tc.infoNotifications,
				}).
				Return(nil)

			err := Progress(template.Execute, lpaStoreResolvingService, progressTracker, certificateProviderStore)(testAppData, w, r, tc.donor)
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

	err := Progress(nil, lpaStoreResolvingService, nil, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetProgressWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{Submitted: true}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything).
		Return(task.Progress{DonorSigned: task.ProgressTask{Done: true}})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := Progress(nil, lpaStoreResolvingService, progressTracker, certificateProviderStore)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
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

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := Progress(template.Execute, lpaStoreResolvingService, progressTracker, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}
