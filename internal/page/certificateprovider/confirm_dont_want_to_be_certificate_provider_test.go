package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

	assert.Nil(t, err)
}

func TestGetConfirmDontWantToBeCertificateProviderWhenShareCodeStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?referenceNumber=123", nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(actor.ShareCodeData{}, expectedError)

	err := ConfirmDontWantToBeCertificateProvider(nil, shareCodeStore, nil, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmDontWantToBeCertificateProviderWhenLpaResolvingServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?referenceNumber=123", nil)

	lpa := lpastore.Lpa{}

	shareCodeData := actor.ShareCodeData{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(shareCodeData, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpa, expectedError)

	err := ConfirmDontWantToBeCertificateProvider(nil, shareCodeStore, lpaStoreResolvingService, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmDontWantToBeCertificateProviderWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?referenceNumber=123", nil)

	lpa := lpastore.Lpa{LpaUID: "lpa-uid"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	shareCodeData := actor.ShareCodeData{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(shareCodeData, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpa, nil)

	err := ConfirmDontWantToBeCertificateProvider(template.Execute, shareCodeStore, lpaStoreResolvingService, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}
