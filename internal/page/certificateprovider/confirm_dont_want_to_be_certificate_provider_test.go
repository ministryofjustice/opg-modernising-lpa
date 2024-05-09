package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?LpaID=lpa-id", nil)

	lpa := lpastore.Lpa{LpaUID: "lpa-uid"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeCertificateProviderData{
			App: testAppData,
			Lpa: &lpa,
		}).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(&lpa, nil)

	err := ConfirmDontWantToBeCertificateProvider(template.Execute, nil, lpaStoreResolvingService, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeCertificateProviderErrors(t *testing.T) {
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
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?LpaID=lpa-id", nil)

			err := ConfirmDontWantToBeCertificateProvider(tc.template().Execute, nil, tc.lpaStoreResolvingService(), nil, nil)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderSignedIn(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?LpaID=lpa-id", nil)
	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})

	testcases := map[string]struct {
		lpa            lpastore.Lpa
		lpaStoreClient func() *mockLpaStoreClient
		donorStore     func() *mockDonorStore
	}{
		"signed in - witnessed and signed": {
			lpa: lpastore.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now(), Donor: actor.Donor{FirstNames: "a b", LastName: "c"}},
			lpaStoreClient: func() *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(ctx, "lpa-uid").
					Return(nil)

				return lpaStoreClient
			},
			donorStore: func() *mockDonorStore { return nil },
		},
		"signed in - not witnessed and signed": {
			lpa:            lpastore.Lpa{LpaUID: "lpa-uid", Donor: actor.Donor{FirstNames: "a b", LastName: "c"}},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
					Return(&actor.DonorProvidedDetails{
						Tasks: actor.DonorTasks{
							CertificateProvider: actor.TaskCompleted,
							CheckYourLpa:        actor.TaskCompleted,
						},
						CertificateProvider: actor.CertificateProvider{UID: actoruid.New()},
					}, nil)
				donorStore.EXPECT().
					Put(ctx, &actor.DonorProvidedDetails{
						Tasks: actor.DonorTasks{
							CertificateProvider: actor.TaskNotStarted,
							CheckYourLpa:        actor.TaskNotStarted,
						},
						CertificateProvider: actor.CertificateProvider{},
					}).
					Return(nil)

				return donorStore
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(ctx).
				Return(&tc.lpa, nil)

			err := ConfirmDontWantToBeCertificateProvider(nil, nil, lpaStoreResolvingService, tc.lpaStoreClient(), tc.donorStore())(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.Paths.CertificateProvider.YouHaveDecidedNotToBeACertificateProvider.Format()+"?donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderNotSignedIn(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123&LpaID=lpa-id", nil)
	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})

	testcases := map[string]struct {
		lpa            lpastore.Lpa
		lpaStoreClient func() *mockLpaStoreClient
		donorStore     func() *mockDonorStore
	}{
		"witnessed and signed": {
			lpa: lpastore.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now(), Donor: actor.Donor{FirstNames: "a b", LastName: "c"}},
			lpaStoreClient: func() *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(ctx, "lpa-uid").
					Return(nil)

				return lpaStoreClient
			},
			donorStore: func() *mockDonorStore { return nil },
		},
		"not witnessed and signed": {
			lpa:            lpastore.Lpa{LpaUID: "lpa-uid", Donor: actor.Donor{FirstNames: "a b", LastName: "c"}},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
					Return(&actor.DonorProvidedDetails{
						Tasks: actor.DonorTasks{
							CertificateProvider: actor.TaskCompleted,
							CheckYourLpa:        actor.TaskCompleted,
						},
						CertificateProvider: actor.CertificateProvider{UID: actoruid.New()},
					}, nil)
				donorStore.EXPECT().
					Put(ctx, &actor.DonorProvidedDetails{
						Tasks: actor.DonorTasks{
							CertificateProvider: actor.TaskNotStarted,
							CheckYourLpa:        actor.TaskNotStarted,
						},
						CertificateProvider: actor.CertificateProvider{},
					}).
					Return(nil)

				return donorStore
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

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

			err := ConfirmDontWantToBeCertificateProvider(nil, shareCodeStore, lpaStoreResolvingService, tc.lpaStoreClient(), tc.donorStore())(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.Paths.CertificateProvider.YouHaveDecidedNotToBeACertificateProvider.Format()+"?donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123&LpaID=lpa-id", nil)
	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})

	shareCodeData := actor.ShareCodeData{
		LpaKey: dynamo.LpaKey("lpa-id"),
	}

	unsignedLPA := lpastore.Lpa{LpaUID: "lpa-uid"}
	signedLPA := lpastore.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now()}

	testcases := map[string]struct {
		lpaStoreResolvingService func() *mockLpaStoreResolvingService
		lpaStoreClient           func() *mockLpaStoreClient
		shareCodeStore           func() *mockShareCodeStore
		donorStore               func() *mockDonorStore
	}{
		"when lpaStoreResolvingService error": {
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&lpastore.Lpa{}, expectedError)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			shareCodeStore: func() *mockShareCodeStore { return nil },
			donorStore:     func() *mockDonorStore { return nil },
		},
		"when lpaStoreClient error": {
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
					SendCertificateProviderOptOut(mock.Anything, mock.Anything).
					Return(expectedError)

				return lpaStoreClient
			},
			shareCodeStore: func() *mockShareCodeStore { return nil },
			donorStore:     func() *mockDonorStore { return nil },
		},
		"when donorStore.GetAny() error": {
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&unsignedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			shareCodeStore: func() *mockShareCodeStore { return nil },
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					GetAny(ctx).
					Return(&actor.DonorProvidedDetails{}, expectedError)

				return donorStore
			},
		},
		"when donorStore.Put() error": {
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&unsignedLPA, nil)

				return lpaStoreResolvingService
			},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
			shareCodeStore: func() *mockShareCodeStore { return nil },
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
		},
		"when shareCodeStore.Get() error": {
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
					SendCertificateProviderOptOut(mock.Anything, mock.Anything).
					Return(nil)

				return lpaStoreClient
			},
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, expectedError)

				return shareCodeStore
			},
			donorStore: func() *mockDonorStore { return nil },
		},
		"when shareCodeStore.Delete() error": {
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
					SendCertificateProviderOptOut(mock.Anything, mock.Anything).
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
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := ConfirmDontWantToBeCertificateProvider(nil, tc.shareCodeStore(), tc.lpaStoreResolvingService(), tc.lpaStoreClient(), tc.donorStore())(testAppData, w, r)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
