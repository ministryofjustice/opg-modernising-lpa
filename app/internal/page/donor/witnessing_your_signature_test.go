package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var now = time.Now()

func TestGetWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingYourSignatureData{App: testAppData, Lpa: lpa}).
		Return(nil)

	err := WitnessingYourSignature(template.Execute, nil)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingYourSignatureWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &witnessingYourSignatureData{App: testAppData, Lpa: lpa}).
		Return(expectedError)

	err := WitnessingYourSignature(template.Execute, nil)(testAppData, w, r, lpa)

	assert.Equal(t, expectedError, err)
}

func TestPostWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpa := &page.Lpa{
		DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin},
		CertificateProvider:   actor.CertificateProvider{Mobile: "07535111111"},
	}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("Send", r.Context(), lpa, mock.Anything).
		Return(nil)

	err := WitnessingYourSignature(nil, witnessCodeSender)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.WitnessingAsCertificateProvider, resp.Header.Get("Location"))
}

func TestPostWitnessingYourSignatureWhenWitnessCodeSenderErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("Send", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := WitnessingYourSignature(nil, witnessCodeSender)(testAppData, w, r, lpa)

	assert.Equal(t, expectedError, err)
}
