package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCertificateProviderSummary(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		LpaUID: "lpa-uid",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(r.Context(), "lpa-uid").
		Return(nil, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &certificateProviderSummaryData{
			App:            testAppData,
			Donor:          donor,
			CanChangeEmail: true,
		}).
		Return(nil)

	err := CertificateProviderSummary(template.Execute, certificateProviderStore)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCertificateProviderSummaryWhenCertificateProviderExists(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		LpaUID: "lpa-uid",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(r.Context(), "lpa-uid").
		Return(nil, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &certificateProviderSummaryData{
			App:   testAppData,
			Donor: donor,
		}).
		Return(nil)

	err := CertificateProviderSummary(template.Execute, certificateProviderStore)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCertificateProviderSummaryWhenCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		LpaUID: "lpa-uid",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := CertificateProviderSummary(nil, certificateProviderStore)(testAppData, w, r, donor)
	assert.ErrorIs(t, err, expectedError)
}

func TestCertificateProviderSummaryWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := CertificateProviderSummary(template.Execute, certificateProviderStore)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}
