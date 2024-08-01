package certificateproviderpage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := lpastore.Lpa{LpaUID: "lpa-uid"}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeCertificateProviderData{
			App: testAppData,
			Lpa: &lpa,
		}).
		Return(nil)

	err := ConfirmDontWantToBeCertificateProvider(template.Execute, lpaStoreResolvingService, nil, nil, nil, nil, "example.com")(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeCertificateProviderErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	testcases := map[string]struct {
		lpaStoreResolvingService func() *mockLpaStoreResolvingService
		template                 func() *mockTemplate
	}{
		"when lpaStoreResolvingService error": {
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(mock.Anything).
					Return(&lpastore.Lpa{}, expectedError)

				return lpaStoreResolvingService
			},
			template: func() *mockTemplate { return nil },
		},
		"when template error": {
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(mock.Anything).
					Return(&lpastore.Lpa{}, nil)

				return lpaStoreResolvingService
			},
			template: func() *mockTemplate {
				template := newMockTemplate(t)
				template.EXPECT().
					Execute(mock.Anything, mock.Anything).
					Return(expectedError)

				return template
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			err := ConfirmDontWantToBeCertificateProvider(tc.template().Execute, tc.lpaStoreResolvingService(), nil, nil, nil, nil, "example.com")(testAppData, w, r, nil)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProvider(t *testing.T) {
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123", SessionID: "456"}), http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()
	uid := actoruid.New()

	testcases := map[string]struct {
		lpa            lpastore.Lpa
		lpaStoreClient func() *mockLpaStoreClient
		donorStore     func() *mockDonorStore
		email          notify.Email
	}{
		"witnessed and signed": {
			lpa: lpastore.Lpa{
				LpaUID:   "lpa-uid",
				SignedAt: time.Now(),
				Donor: lpastore.Donor{
					FirstNames: "a b", LastName: "c", Email: "a@example.com",
				},
				CertificateProvider: lpastore.CertificateProvider{
					FirstNames: "d e", LastName: "f", UID: uid,
				},
				Type: actor.LpaTypePersonalWelfare,
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
				CertificateProviderFirstNames: "d e",
				CertificateProviderFullName:   "d e f",
				DonorFullName:                 "a b c",
				LpaType:                       "Personal welfare",
				LpaUID:                        "lpa-uid",
				DonorStartPageURL:             "example.com" + page.Paths.Start.Format(),
			},
		},
		"cannot-register": {
			lpa: lpastore.Lpa{
				LpaUID:   "lpa-uid",
				SignedAt: time.Now(),
				Donor: lpastore.Donor{
					FirstNames: "a b", LastName: "c", Email: "a@example.com",
				},
				CertificateProvider: lpastore.CertificateProvider{
					FirstNames: "d e", LastName: "f", UID: uid,
				},
				Type:           actor.LpaTypePersonalWelfare,
				CannotRegister: true,
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore:     func() *mockDonorStore { return nil },
			email: notify.CertificateProviderOptedOutPostWitnessingEmail{
				CertificateProviderFirstNames: "d e",
				CertificateProviderFullName:   "d e f",
				DonorFullName:                 "a b c",
				LpaType:                       "Personal welfare",
				LpaUID:                        "lpa-uid",
				DonorStartPageURL:             "example.com" + page.Paths.Start.Format(),
			},
		},
		"not witnessed and signed": {
			lpa: lpastore.Lpa{
				LpaUID: "lpa-uid",
				Donor:  lpastore.Donor{FirstNames: "a b", LastName: "c", Email: "a@example.com"},
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(r.Context()).
					Return(&actor.DonorProvidedDetails{
						LpaUID: "lpa-uid",
						Donor: actor.Donor{
							FirstNames: "a b", LastName: "c",
						},
						Tasks: actor.DonorTasks{
							CertificateProvider: actor.TaskCompleted,
							CheckYourLpa:        actor.TaskCompleted,
						},
						CertificateProvider: actor.CertificateProvider{
							UID:        uid,
							FirstNames: "d e", LastName: "f",
						},
						Type: actor.LpaTypePersonalWelfare,
					}, nil)
				donorStore.EXPECT().
					Put(r.Context(), &actor.DonorProvidedDetails{
						LpaUID: "lpa-uid",
						Donor: actor.Donor{
							FirstNames: "a b", LastName: "c",
						},
						Tasks: actor.DonorTasks{
							CertificateProvider: actor.TaskNotStarted,
							CheckYourLpa:        actor.TaskNotStarted,
						},
						CertificateProvider: actor.CertificateProvider{},
						Type:                actor.LpaTypePersonalWelfare,
					}).
					Return(nil)

				return donorStore
			},
			email: notify.CertificateProviderOptedOutPreWitnessingEmail{
				CertificateProviderFullName: "d e f",
				DonorFullName:               "a b c",
				LpaType:                     "Personal welfare",
				LpaUID:                      "lpa-uid",
				DonorStartPageURL:           "example.com" + page.Paths.Start.Format(),
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&tc.lpa, nil)

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
				SendActorEmail(r.Context(), "a@example.com", "lpa-uid", tc.email).
				Return(nil)

			err := ConfirmDontWantToBeCertificateProvider(nil, lpaStoreResolvingService, tc.lpaStoreClient(), tc.donorStore(), certificateProviderStore, notifyClient, "example.com")(testAppData, w, r, nil)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.Paths.CertificateProvider.YouHaveDecidedNotToBeCertificateProvider.Format()+"?donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)

	unsignedLPA := lpastore.Lpa{LpaUID: "lpa-uid"}
	signedLPA := lpastore.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now()}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("a")

	testcases := map[string]struct {
		sessionStore             func() *mockSessionStore
		lpaStoreResolvingService func() *mockLpaStoreResolvingService
		lpaStoreClient           func() *mockLpaStoreClient
		donorStore               func() *mockDonorStore
		certificateProviderStore func() *mockCertificateProviderStore
		localizer                func() *mockLocalizer
		notifyClient             func() *mockNotifyClient
	}{
		"when lpaStoreClient error": {
			sessionStore: func() *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(r.Context()).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return lpaStoreClient
			},
			donorStore:               func() *mockDonorStore { return nil },
			certificateProviderStore: func() *mockCertificateProviderStore { return nil },
			localizer:                func() *mockLocalizer { return localizer },
			notifyClient:             func() *mockNotifyClient { return nil },
		},
		"when donorStore.GetAny() error": {
			sessionStore: func() *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(r.Context()).
					Return(&unsignedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(r.Context()).
					Return(&actor.DonorProvidedDetails{}, expectedError)

				return donorStore
			},
			certificateProviderStore: func() *mockCertificateProviderStore { return nil },
			localizer:                func() *mockLocalizer { return nil },
			notifyClient:             func() *mockNotifyClient { return nil },
		},
		"when donorStore.Put() error": {
			sessionStore: func() *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(r.Context()).
					Return(&unsignedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(r.Context()).
					Return(&actor.DonorProvidedDetails{}, nil)
				donorStore.EXPECT().
					Put(r.Context(), mock.Anything).
					Return(expectedError)

				return donorStore
			},
			certificateProviderStore: func() *mockCertificateProviderStore { return nil },
			localizer:                func() *mockLocalizer { return localizer },
			notifyClient:             func() *mockNotifyClient { return nil },
		},
		"when certificateProviderStore.Delete() error": {
			sessionStore: func() *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(r.Context()).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return lpaStoreClient
			},
			donorStore: func() *mockDonorStore { return nil },
			certificateProviderStore: func() *mockCertificateProviderStore {
				certificateProviderStore := newMockCertificateProviderStore(t)
				certificateProviderStore.EXPECT().
					Delete(mock.Anything).
					Return(expectedError)

				return certificateProviderStore
			},
			localizer:    func() *mockLocalizer { return localizer },
			notifyClient: func() *mockNotifyClient { return nil },
		},
		"when notifyClient.SendActorEmail() error": {
			sessionStore: func() *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(r.Context()).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return lpaStoreClient
			},
			donorStore: func() *mockDonorStore { return nil },
			certificateProviderStore: func() *mockCertificateProviderStore {
				certificateProviderStore := newMockCertificateProviderStore(t)
				certificateProviderStore.EXPECT().
					Delete(mock.Anything).
					Return(nil)

				return certificateProviderStore
			},
			localizer: func() *mockLocalizer { return localizer },
			notifyClient: func() *mockNotifyClient {
				client := newMockNotifyClient(t)
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

			testAppData.Localizer = tc.localizer()

			err := ConfirmDontWantToBeCertificateProvider(nil, tc.lpaStoreResolvingService(), tc.lpaStoreClient(), tc.donorStore(), tc.certificateProviderStore(), tc.notifyClient(), "example.com")(testAppData, w, r, nil)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
