package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeCertificateProviderLoggedOut(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpadata.Lpa{LpaUID: "lpa-uid"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeCertificateProviderDataLoggedOut{
			App: testAppData,
			Lpa: lpa,
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
					Return(&lpadata.Lpa{}, expectedError)

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
					Return(&lpadata.Lpa{}, nil)

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
	r, _ := http.NewRequest(http.MethodPost, "/?code=da4ec3358a10c9b0872eb877953cc7b07af5f4d75e4c1cb0597cbbf41e5dbe35", nil)
	w := httptest.NewRecorder()
	ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})

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
					FirstNames: "d e", LastName: "f",
				},
				Type: lpadata.LpaTypePersonalWelfare,
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
				Donor:                            lpadata.Donor{FirstNames: "a b", LastName: "c", Email: "a@example.com", ContactLanguagePreference: localize.En},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "d e", LastName: "f",
				},
				Status: lpadata.StatusCannotRegister,
				Type:   lpadata.LpaTypePersonalWelfare,
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
				Donor: lpadata.Donor{
					FirstNames: "a b", LastName: "c", Email: "a@example.com",
					ContactLanguagePreference: localize.En,
				},
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
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
							UID:        actoruid.New(),
							FirstNames: "d e", LastName: "f",
						},
						Type: lpadata.LpaTypePersonalWelfare,
					}, nil)
				donorStore.EXPECT().
					Put(ctx, &donordata.Provided{
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
			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				LpaData(r).
				Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

			accessCodeData := accesscodedata.Link{
				LpaKey:      dynamo.LpaKey("lpa-id"),
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			}

			accessCodeStore := newMockAccessCodeStore(t)
			accessCodeStore.EXPECT().
				Get(r.Context(), actor.TypeCertificateProvider, accesscodedata.HashedFromString("abcdef123456")).
				Return(accessCodeData, nil)
			accessCodeStore.EXPECT().
				Delete(r.Context(), accessCodeData).
				Return(nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(ctx).
				Return(tc.lpa, nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				EmailGreeting(tc.lpa).
				Return("Dear donor")
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaDonor(tc.lpa), "lpa-uid", tc.email).
				Return(nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("personal-welfare").
				Return("Personal welfare")

			testAppData.Localizer = localizer

			err := ConfirmDontWantToBeCertificateProviderLoggedOut(nil, accessCodeStore, lpaStoreResolvingService, tc.lpaStoreClient(), tc.donorStore(), sessionStore, notifyClient, "example.com")(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.PathCertificateProviderYouHaveDecidedNotToBeCertificateProvider.Format()+"?donorFirstNames=a+b&donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderLoggedOutErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})

	accessCodeData := accesscodedata.Link{
		LpaKey: dynamo.LpaKey("lpa-id"),
	}

	unsignedLPA := lpadata.Lpa{LpaUID: "lpa-uid"}
	signedLPA := lpadata.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now(), WitnessedByCertificateProviderAt: time.Now()}
	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("a")

	testcases := map[string]struct {
		sessionStore             func(*testing.T) *mockSessionStore
		lpaStoreResolvingService func(*testing.T) *mockLpaStoreResolvingService
		lpaStoreClient           func(*testing.T) *mockLpaStoreClient
		accessCodeStore          func(*testing.T) *mockAccessCodeStore
		donorStore               func(*testing.T) *mockDonorStore
		localizer                func(*testing.T) *mockLocalizer
		notifyClient             func(*testing.T) *mockNotifyClient
	}{
		"when lpaStoreClient error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return lpaStoreClient
			},
			accessCodeStore: func(t *testing.T) *mockAccessCodeStore {
				accessCodeStore := newMockAccessCodeStore(t)
				accessCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(accessCodeData, nil)

				return accessCodeStore
			},
			donorStore: func(t *testing.T) *mockDonorStore { return nil },
			localizer:  func(t *testing.T) *mockLocalizer { return localizer },
			notifyClient: func(t *testing.T) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")

				return notifyClient
			},
		},
		"when donorStore.GetAny() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&unsignedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient { return nil },
			accessCodeStore: func(t *testing.T) *mockAccessCodeStore {
				accessCodeStore := newMockAccessCodeStore(t)
				accessCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(accessCodeData, nil)

				return accessCodeStore
			},
			donorStore: func(t *testing.T) *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
					Return(&donordata.Provided{}, expectedError)

				return donorStore
			},
			localizer:    func(t *testing.T) *mockLocalizer { return nil },
			notifyClient: func(t *testing.T) *mockNotifyClient { return nil },
		},
		"when donorStore.Put() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&unsignedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient { return nil },
			accessCodeStore: func(t *testing.T) *mockAccessCodeStore {
				accessCodeStore := newMockAccessCodeStore(t)
				accessCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(accessCodeData, nil)

				return accessCodeStore
			},
			donorStore: func(t *testing.T) *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
					Return(&donordata.Provided{}, nil)
				donorStore.EXPECT().
					Put(ctx, mock.Anything).
					Return(expectedError)

				return donorStore
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
		"when accessCodeStore.Get() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient { return nil },
			accessCodeStore: func(t *testing.T) *mockAccessCodeStore {
				accessCodeStore := newMockAccessCodeStore(t)
				accessCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(accessCodeData, expectedError)

				return accessCodeStore
			},
			donorStore:   func(t *testing.T) *mockDonorStore { return nil },
			localizer:    func(t *testing.T) *mockLocalizer { return localizer },
			notifyClient: func(t *testing.T) *mockNotifyClient { return nil },
		},
		"when accessCodeStore.Delete() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return lpaStoreClient
			},
			accessCodeStore: func(t *testing.T) *mockAccessCodeStore {
				accessCodeStore := newMockAccessCodeStore(t)
				accessCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(accessCodeData, nil)
				accessCodeStore.EXPECT().
					Delete(mock.Anything, mock.Anything).
					Return(expectedError)

				return accessCodeStore
			},
			donorStore: func(t *testing.T) *mockDonorStore { return nil },
			localizer:  func(t *testing.T) *mockLocalizer { return localizer },
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return client
			},
		},
		"when notifyClient.SendActorEmail() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return lpaStoreClient
			},
			accessCodeStore: func(t *testing.T) *mockAccessCodeStore {
				accessCodeStore := newMockAccessCodeStore(t)
				accessCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(accessCodeData, nil)

				return accessCodeStore
			},
			donorStore: func(t *testing.T) *mockDonorStore { return nil },
			localizer:  func(t *testing.T) *mockLocalizer { return localizer },
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("")
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

			err := ConfirmDontWantToBeCertificateProviderLoggedOut(nil, tc.accessCodeStore(t), tc.lpaStoreResolvingService(t), tc.lpaStoreClient(t), tc.donorStore(t), tc.sessionStore(t), tc.notifyClient(t), "example.com")(testAppData, w, r)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
