package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeCertificateProviderLoggedOut(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := lpastore.Lpa{LpaUID: "lpa-uid"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(&lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeCertificateProviderDataLoggedOut{
			App: testAppData,
			Lpa: &lpa,
		}).
		Return(nil)

	err := ConfirmDontWantToBeCertificateProviderLoggedOut(template.Execute, nil, lpaStoreResolvingService, nil, nil, sessionStore, nil, "example.com")(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeCertificateProviderLoggedOutErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	testcases := map[string]struct {
		sessionStore             func() *mockSessionStore
		lpaStoreResolvingService func() *mockLpaStoreResolvingService
		template                 func() *mockTemplate
	}{
		"when sessionStore error": {
			sessionStore: func() *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{}, expectedError)

				return sessionStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService { return nil },
			template:                 func() *mockTemplate { return nil },
		},
		"when lpaStoreResolvingService error": {
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
					Get(mock.Anything).
					Return(&lpastore.Lpa{}, expectedError)

				return lpaStoreResolvingService
			},
			template: func() *mockTemplate { return nil },
		},
		"when template error": {
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
			err := ConfirmDontWantToBeCertificateProviderLoggedOut(tc.template().Execute, nil, tc.lpaStoreResolvingService(), nil, nil, tc.sessionStore(), nil, "example.com")(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderLoggedOut(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()
	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})

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
					FirstNames: "d e", LastName: "f",
				},
				Type: actor.LpaTypePersonalWelfare,
			},
			lpaStoreClient: func() *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(ctx, "lpa-uid", actoruid.Service).
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
				Donor:    lpastore.Donor{FirstNames: "a b", LastName: "c", Email: "a@example.com"},
				CertificateProvider: lpastore.CertificateProvider{
					FirstNames: "d e", LastName: "f",
				},
				CannotRegister: true,
				Type:           actor.LpaTypePersonalWelfare,
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
				Donor: lpastore.Donor{
					FirstNames: "a b", LastName: "c", Email: "a@example.com",
				},
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
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
							UID:        actoruid.New(),
							FirstNames: "d e", LastName: "f",
						},
						Type: actor.LpaTypePersonalWelfare,
					}, nil)
				donorStore.EXPECT().
					Put(ctx, &actor.DonorProvidedDetails{
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
			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				LpaData(r).
				Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

			shareCodeData := actor.ShareCodeData{
				LpaKey:      dynamo.LpaKey("lpa-id"),
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			}

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Get(r.Context(), actor.TypeCertificateProvider, "123").
				Return(shareCodeData, nil)
			shareCodeStore.EXPECT().
				Delete(r.Context(), shareCodeData).
				Return(nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(ctx).
				Return(&tc.lpa, nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, "a@example.com", "lpa-uid", tc.email).
				Return(nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("personal-welfare").
				Return("Personal welfare")

			testAppData.Localizer = localizer

			err := ConfirmDontWantToBeCertificateProviderLoggedOut(nil, shareCodeStore, lpaStoreResolvingService, tc.lpaStoreClient(), tc.donorStore(), sessionStore, notifyClient, "example.com")(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.Paths.CertificateProvider.YouHaveDecidedNotToBeCertificateProvider.Format()+"?donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderLoggedOutErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})

	shareCodeData := actor.ShareCodeData{
		LpaKey: dynamo.LpaKey("lpa-id"),
	}

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
		shareCodeStore           func() *mockShareCodeStore
		donorStore               func() *mockDonorStore
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
					Get(ctx).
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
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
			},
			donorStore:   func() *mockDonorStore { return nil },
			localizer:    func() *mockLocalizer { return localizer },
			notifyClient: func() *mockNotifyClient { return nil },
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
					Get(ctx).
					Return(&unsignedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
			},
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
					Return(&actor.DonorProvidedDetails{}, expectedError)

				return donorStore
			},
			localizer:    func() *mockLocalizer { return nil },
			notifyClient: func() *mockNotifyClient { return nil },
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
					Get(ctx).
					Return(&unsignedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
			},
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
					Return(&actor.DonorProvidedDetails{}, nil)
				donorStore.EXPECT().
					Put(ctx, mock.Anything).
					Return(expectedError)

				return donorStore
			},
			localizer:    func() *mockLocalizer { return localizer },
			notifyClient: func() *mockNotifyClient { return nil },
		},
		"when shareCodeStore.Get() error": {
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
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, expectedError)

				return shareCodeStore
			},
			donorStore:   func() *mockDonorStore { return nil },
			localizer:    func() *mockLocalizer { return localizer },
			notifyClient: func() *mockNotifyClient { return nil },
		},
		"when shareCodeStore.Delete() error": {
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
					Get(ctx).
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
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)
				shareCodeStore.EXPECT().
					Delete(mock.Anything, mock.Anything).
					Return(expectedError)

				return shareCodeStore
			},
			donorStore: func() *mockDonorStore { return nil },
			localizer:  func() *mockLocalizer { return localizer },
			notifyClient: func() *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return client
			},
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
					Get(ctx).
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
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
			},
			donorStore: func() *mockDonorStore { return nil },
			localizer:  func() *mockLocalizer { return localizer },
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

			err := ConfirmDontWantToBeCertificateProviderLoggedOut(nil, tc.shareCodeStore(), tc.lpaStoreResolvingService(), tc.lpaStoreClient(), tc.donorStore(), tc.sessionStore(), tc.notifyClient(), "example.com")(testAppData, w, r)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
