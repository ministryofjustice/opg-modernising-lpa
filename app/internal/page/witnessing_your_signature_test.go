package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var now = time.Now()

type mockNotifyClient struct {
	mock.Mock
}

func (m *mockNotifyClient) TemplateID(name string) string {
	return m.Called(name).String(0)
}

func (m *mockNotifyClient) Email(ctx context.Context, email notify.Email) (string, error) {
	args := m.Called(ctx, email)
	return args.String(0), args.Error(1)
}

func (m *mockNotifyClient) Sms(ctx context.Context, sms notify.Sms) (string, error) {
	args := m.Called(ctx, sms)
	return args.String(0), args.Error(1)
}

func TestGetWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{CertificateProvider: CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingYourSignatureData{App: appData, Lpa: lpa}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WitnessingYourSignature(template.Func, lpaStore, nil, nil, nil)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetWitnessingYourSignatureWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WitnessingYourSignature(nil, lpaStore, nil, nil, nil)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetWitnessingYourSignatureWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{CertificateProvider: CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &witnessingYourSignatureData{App: appData, Lpa: lpa}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WitnessingYourSignature(template.Func, lpaStore, nil, nil, nil)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostWitnessingYourSignature(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{CertificateProvider: CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			CertificateProvider: CertificateProvider{Mobile: "07535111111"},
			WitnessCode: WitnessCode{
				Code:    "1234",
				Created: now,
			},
			SignatureSmsID: "sms-id",
		}).
		Return(nil)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", "MLPA Beta signature code").
		Return("xyz")
	notifyClient.
		On("Sms", mock.Anything, notify.Sms{
			PhoneNumber:     "07535111111",
			TemplateID:      "xyz",
			Personalisation: map[string]string{"code": "1234"},
		}).
		Return("sms-id", nil)

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := WitnessingYourSignature(nil, lpaStore, notifyClient, func(l int) string { return "1234" }, func() time.Time { return now })(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.WitnessingAsCertificateProvider, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
}

func TestPostWitnessingYourSignatureWhenNotifyErrors(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{CertificateProvider: CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", "MLPA Beta signature code").
		Return("xyz")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("", expectedError)

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := WitnessingYourSignature(nil, lpaStore, notifyClient, func(l int) string { return "1234" }, func() time.Time { return now })(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
}

func TestPostWitnessingYourSignatureWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{CertificateProvider: CertificateProvider{Mobile: "07535111111"}}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", "MLPA Beta signature code").
		Return("xyz")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("sms-id", nil)

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := WitnessingYourSignature(nil, lpaStore, notifyClient, func(l int) string { return "1234" }, func() time.Time { return now })(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
}
