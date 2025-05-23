package certificateproviderpage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpadata.Lpa{LpaUID: "lpa-uid"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeCertificateProviderData{
			App: testAppData,
			Lpa: lpa,
		}).
		Return(nil)

	err := ConfirmDontWantToBeCertificateProvider(template.Execute, nil, nil, nil, nil, "example.com")(testAppData, w, r, nil, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ConfirmDontWantToBeCertificateProvider(template.Execute, nil, nil, nil, nil, "example.com")(testAppData, w, r, nil, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmDontWantToBeCertificateProvider(t *testing.T) {
	r, _ := http.NewRequestWithContext(appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"}), http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()
	uid := actoruid.New()

	testcases := map[string]struct {
		lpa            *lpadata.Lpa
		lpaStoreClient func() *mockLpaStoreClient
		donorStore     func() *mockDonorStore
		email          notify.Email
	}{
		"witnessed and signed": {
			lpa: &lpadata.Lpa{
				LpaUID:                           "lpa-uid",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Donor: lpadata.Donor{
					FirstNames: "a b", LastName: "c", Email: "a@example.com",
					ContactLanguagePreference: localize.En,
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "d e", LastName: "f", UID: uid,
				},
				Type: lpadata.LpaTypePersonalWelfare,
			},
			lpaStoreClient: func() *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(r.Context(), "lpa-uid", uid).
					Return(nil)

				return lpaStoreClient
			},
			donorStore: func() *mockDonorStore { return nil },
			email: notify.CertificateProviderOptedOutPostWitnessingEmail{
				Greeting:                      "Dear donor",
				CertificateProviderFirstNames: "d e",
				CertificateProviderFullName:   "d e f",
				DonorFullName:                 "a b c",
				LpaType:                       "Personal welfare",
				LpaReferenceNumber:            "lpa-uid",
				DonorStartPageURL:             "example.com",
			},
		},
		"cannot-register": {
			lpa: &lpadata.Lpa{
				LpaUID:                           "lpa-uid",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Donor: lpadata.Donor{
					FirstNames: "a b", LastName: "c", Email: "a@example.com",
					ContactLanguagePreference: localize.En,
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "d e", LastName: "f", UID: uid,
				},
				Type:   lpadata.LpaTypePersonalWelfare,
				Status: lpadata.StatusCannotRegister,
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore:     func() *mockDonorStore { return nil },
			email: notify.CertificateProviderOptedOutPostWitnessingEmail{
				Greeting:                      "Dear donor",
				CertificateProviderFirstNames: "d e",
				CertificateProviderFullName:   "d e f",
				DonorFullName:                 "a b c",
				LpaType:                       "Personal welfare",
				LpaReferenceNumber:            "lpa-uid",
				DonorStartPageURL:             "example.com",
			},
		},
		"not witnessed and signed": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Donor:  lpadata.Donor{FirstNames: "a b", LastName: "c", Email: "a@example.com", ContactLanguagePreference: localize.En},
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(r.Context()).
					Return(&donordata.Provided{
						LpaUID: "lpa-uid",
						Donor: donordata.Donor{
							FirstNames: "a b", LastName: "c",
						},
						Tasks: donordata.Tasks{
							CertificateProvider: task.StateCompleted,
							CheckYourLpa:        task.StateCompleted,
						},
						CertificateProvider: donordata.CertificateProvider{
							UID:        uid,
							FirstNames: "d e", LastName: "f",
						},
						Type: lpadata.LpaTypePersonalWelfare,
					}, nil)
				donorStore.EXPECT().
					Put(r.Context(), &donordata.Provided{
						LpaUID: "lpa-uid",
						Donor: donordata.Donor{
							FirstNames: "a b", LastName: "c",
						},
						Tasks: donordata.Tasks{
							CertificateProvider: task.StateNotStarted,
							CheckYourLpa:        task.StateNotStarted,
						},
						CertificateProvider: donordata.CertificateProvider{},
						Type:                lpadata.LpaTypePersonalWelfare,
					}).
					Return(nil)

				return donorStore
			},
			email: notify.CertificateProviderOptedOutPreWitnessingEmail{
				Greeting:                    "Dear donor",
				CertificateProviderFullName: "d e f",
				DonorFullName:               "a b c",
				LpaType:                     "Personal welfare",
				LpaReferenceNumber:          "lpa-uid",
				DonorStartPageURL:           "example.com",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				Delete(r.Context()).
				Return(nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("personal-welfare").
				Return("Personal welfare")

			testAppData.Localizer = localizer

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				EmailGreeting(mock.Anything).
				Return("Dear donor")
			notifyClient.EXPECT().
				SendActorEmail(r.Context(), notify.ToLpaDonor(tc.lpa), "lpa-uid", tc.email).
				Return(nil)

			err := ConfirmDontWantToBeCertificateProvider(nil, tc.lpaStoreClient(), tc.donorStore(), certificateProviderStore, notifyClient, "example.com")(testAppData, w, r, nil, tc.lpa)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.PathCertificateProviderYouHaveDecidedNotToBeCertificateProvider.Format()+"?donorFirstNames=a+b&donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)

	unsignedLPA := &lpadata.Lpa{LpaUID: "lpa-uid"}
	signedLPA := &lpadata.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now(), WitnessedByCertificateProviderAt: time.Now()}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("a")

	testcases := map[string]struct {
		lpa                      *lpadata.Lpa
		sessionStore             func(*testing.T) *mockSessionStore
		lpaStoreClient           func(*testing.T) *mockLpaStoreClient
		donorStore               func(*testing.T) *mockDonorStore
		certificateProviderStore func(*testing.T) *mockCertificateProviderStore
		localizer                func(*testing.T) *mockLocalizer
		notifyClient             func(*testing.T) *mockNotifyClient
	}{
		"when lpaStoreClient error": {
			lpa: signedLPA,
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return lpaStoreClient
			},
			donorStore:               func(t *testing.T) *mockDonorStore { return nil },
			certificateProviderStore: func(t *testing.T) *mockCertificateProviderStore { return nil },
			localizer:                func(t *testing.T) *mockLocalizer { return localizer },
			notifyClient: func(t *testing.T) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")

				return notifyClient
			},
		},
		"when donorStore.GetAny() error": {
			lpa: unsignedLPA,
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient { return nil },
			donorStore: func(t *testing.T) *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(r.Context()).
					Return(&donordata.Provided{}, expectedError)

				return donorStore
			},
			certificateProviderStore: func(t *testing.T) *mockCertificateProviderStore { return nil },
			localizer:                func(t *testing.T) *mockLocalizer { return nil },
			notifyClient:             func(t *testing.T) *mockNotifyClient { return nil },
		},
		"when donorStore.Put() error": {
			lpa: unsignedLPA,
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient { return nil },
			donorStore: func(t *testing.T) *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(r.Context()).
					Return(&donordata.Provided{}, nil)
				donorStore.EXPECT().
					Put(r.Context(), mock.Anything).
					Return(expectedError)

				return donorStore
			},
			certificateProviderStore: func(t *testing.T) *mockCertificateProviderStore { return nil },
			localizer:                func(t *testing.T) *mockLocalizer { return localizer },
			notifyClient: func(t *testing.T) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")

				return notifyClient
			},
		},
		"when certificateProviderStore.Delete() error": {
			lpa: signedLPA,
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return lpaStoreClient
			},
			donorStore: func(t *testing.T) *mockDonorStore { return nil },
			certificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				certificateProviderStore := newMockCertificateProviderStore(t)
				certificateProviderStore.EXPECT().
					Delete(mock.Anything).
					Return(expectedError)

				return certificateProviderStore
			},
			localizer: func(t *testing.T) *mockLocalizer { return localizer },
			notifyClient: func(t *testing.T) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")

				return notifyClient
			},
		},
		"when notifyClient.SendActorEmail() error": {
			lpa: signedLPA,
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return lpaStoreClient
			},
			donorStore: func(t *testing.T) *mockDonorStore { return nil },
			certificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				certificateProviderStore := newMockCertificateProviderStore(t)
				certificateProviderStore.EXPECT().
					Delete(mock.Anything).
					Return(nil)

				return certificateProviderStore
			},
			localizer: func(t *testing.T) *mockLocalizer { return localizer },
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			testAppData.Localizer = tc.localizer(t)

			err := ConfirmDontWantToBeCertificateProvider(nil, tc.lpaStoreClient(t), tc.donorStore(t), tc.certificateProviderStore(t), tc.notifyClient(t), "example.com")(testAppData, w, r, nil, tc.lpa)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
