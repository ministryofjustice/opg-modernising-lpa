package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var now = time.Now()

func TestGetWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &witnessingYourSignatureData{App: page.TestAppData, Lpa: lpa}).
		Return(nil)

	err := WitnessingYourSignature(template.Func, lpaStore, nil, nil, nil)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetWitnessingYourSignatureWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := WitnessingYourSignature(nil, lpaStore, nil, nil, nil)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetWitnessingYourSignatureWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &witnessingYourSignatureData{App: page.TestAppData, Lpa: lpa}).
		Return(page.ExpectedError)

	err := WitnessingYourSignature(template.Func, lpaStore, nil, nil, nil)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"},
			WitnessCode: page.WitnessCode{
				Code:    "1234",
				Created: now,
			},
			SignatureSmsID: "sms-id",
		}).
		Return(nil)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", notify.SignatureCodeSms).
		Return("xyz")
	notifyClient.
		On("Sms", mock.Anything, notify.Sms{
			PhoneNumber:     "07535111111",
			TemplateID:      "xyz",
			Personalisation: map[string]string{"code": "1234"},
		}).
		Return("sms-id", nil)

	err := WitnessingYourSignature(nil, lpaStore, notifyClient, func(l int) string { return "1234" }, func() time.Time { return now })(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.WitnessingAsCertificateProvider, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
}

func TestPostWitnessingYourSignatureWhenNotifyErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", notify.SignatureCodeSms).
		Return("xyz")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("", page.ExpectedError)

	err := WitnessingYourSignature(nil, lpaStore, notifyClient, func(l int) string { return "1234" }, func() time.Time { return now })(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
}

func TestPostWitnessingYourSignatureWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpa := &page.Lpa{CertificateProvider: actor.CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(page.ExpectedError)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", notify.SignatureCodeSms).
		Return("xyz")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("sms-id", nil)

	err := WitnessingYourSignature(nil, lpaStore, notifyClient, func(l int) string { return "1234" }, func() time.Time { return now })(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
}
