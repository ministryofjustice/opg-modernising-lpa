package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?referenceNumber=123", nil)

	lpa := lpastore.Lpa{LpaUID: "lpa-uid"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeCertificateProviderData{
			App: testAppData,
			Lpa: &lpa,
		}).
		Return(nil)

	shareCodeData := actor.ShareCodeData{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeCertificateProvider, "123").
		Return(shareCodeData, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: "DONOR#donor", LpaID: "lpa-id"})).
		Return(&lpa, nil)

	err := ConfirmDontWantToBeCertificateProvider(template.Execute, shareCodeStore, lpaStoreResolvingService, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeCertificateProviderErrors(t *testing.T) {
	shareCodeData := actor.ShareCodeData{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	testcases := map[string]struct {
		shareCodeStore           func() *mockShareCodeStore
		lpaStoreResolvingService func() *mockLpaStoreResolvingService
		template                 func() *mockTemplate
	}{
		"when shareCodeStore error": {
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(actor.ShareCodeData{}, expectedError)

				return shareCodeStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService { return nil },
			template:                 func() *mockTemplate { return nil },
		},
		"when lpaStoreResolvingService error": {
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
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
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
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
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?referenceNumber=123", nil)

			err := ConfirmDontWantToBeCertificateProvider(tc.template().Execute, tc.shareCodeStore(), tc.lpaStoreResolvingService(), nil)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProvider(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: "DONOR#donor", LpaID: "lpa-id"})

	testcases := map[string]struct {
		lpa            lpastore.Lpa
		lpaStoreClient func() *mockLpaStoreClient
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
		},
		"not witnessed and signed": {
			lpa:            lpastore.Lpa{LpaUID: "lpa-uid", Donor: actor.Donor{FirstNames: "a b", LastName: "c"}},
			lpaStoreClient: func() *mockLpaStoreClient { return nil },
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

			err := ConfirmDontWantToBeCertificateProvider(nil, shareCodeStore, lpaStoreResolvingService, tc.lpaStoreClient())(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.Paths.CertificateProvider.YouHaveDecidedNotToBeACertificateProvider.Format()+"?donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeCertificateProviderErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: "DONOR#donor", LpaID: "lpa-id"})

	shareCodeData := actor.ShareCodeData{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	lpa := lpastore.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now()}

	testcases := map[string]struct {
		lpaStoreClient func() *mockLpaStoreClient
		shareCodeStore func() *mockShareCodeStore
	}{
		"when lpaStoreClient error": {
			lpaStoreClient: func() *mockLpaStoreClient {
				lpaStoreClient := newMockLpaStoreClient(t)
				lpaStoreClient.EXPECT().
					SendCertificateProviderOptOut(mock.Anything, mock.Anything).
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
		},
		"when shareCodeStore error": {
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
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(ctx).
				Return(&lpa, nil)

			err := ConfirmDontWantToBeCertificateProvider(nil, tc.shareCodeStore(), lpaStoreResolvingService, tc.lpaStoreClient())(testAppData, w, r)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
