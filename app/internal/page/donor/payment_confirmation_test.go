package donor

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentConfirmation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := (&mockPayClient{BaseURL: "http://base.url"}).
		withASuccessfulPayment("abc123", "123456789012")

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", notify.CertificateProviderInviteEmail).
		Return("template-id")
	notifyClient.
		On("Email", r.Context(), notify.Email{
			TemplateID:   "template-id",
			EmailAddress: "certificateprovider@example.com",
			Personalisation: map[string]string{
				"link": fmt.Sprintf("http://app%s?share-code=123", page.Paths.CertificateProviderStart),
			},
		}).
		Return("", nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &paymentConfirmationData{App: testAppData, PaymentReference: "123456789012", Continue: testAppData.Paths.TaskList}).
		Return(nil)

	sessionsStore := (&mockSessionsStore{}).
		withPaySession(r).
		withExpiredPaySession(r, w)

	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r).
		withCompletedPaymentLpaData(r, "abc123", "123456789012")

	dataStore := &mockDataStore{}
	dataStore.
		On("Put", r.Context(), "SHARECODE#123", "#METADATA#123", page.ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
		Return(nil)

	err := PaymentConfirmation(&mockLogger{}, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app", dataStore, mockRandom)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore, payClient, lpaStore, sessionsStore)
}

func TestGetPaymentConfirmationGettingLpaErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := &mockTemplate{}
	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	logger := &mockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to retrieve item from data store using key '%s': %s", "session-id", expectedError.Error())).
		Return(nil)

	err := PaymentConfirmation(logger, template.Func, &mockPayClient{}, nil, lpaStore, &mockSessionsStore{}, "http://app", nil, mockRandom)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetPaymentConfirmationWhenErrorGettingSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := &mockTemplate{}
	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r)

	sessionsStore := &mockSessionsStore{}
	sessionsStore.
		On("Get", r, "pay").
		Return(&sessions.Session{}, expectedError)

	err := PaymentConfirmation(nil, template.Func, &mockPayClient{}, nil, lpaStore, sessionsStore, "http://app", nil, mockRandom)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore)
}

func TestGetPaymentConfirmationWhenErrorGettingPayment(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r)

	sessionsStore := (&mockSessionsStore{}).
		withPaySession(r)

	logger := &mockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to retrieve payment info: %s", expectedError.Error())).
		Return(nil)

	payClient := &mockPayClient{}
	payClient.
		On("GetPayment", "abc123").
		Return(pay.GetPaymentResponse{}, expectedError)

	template := &mockTemplate{}

	err := PaymentConfirmation(logger, template.Func, payClient, nil, lpaStore, sessionsStore, "http://app", nil, mockRandom)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore, logger, payClient)
}

func TestGetPaymentConfirmationWhenErrorPuttingShareCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r)

	sessionsStore := (&mockSessionsStore{}).
		withPaySession(r)

	payClient := (&mockPayClient{}).
		withASuccessfulPayment("abc123", "123456789012")

	dataStore := &mockDataStore{}
	dataStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(nil, template.Func, payClient, nil, lpaStore, sessionsStore, "http://app", dataStore, mockRandom)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore, lpaStore, sessionsStore, payClient)
}

func TestGetPaymentConfirmationWhenErrorSendingEmail(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Email", mock.Anything, mock.Anything).
		Return("", expectedError)

	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r)

	sessionsStore := (&mockSessionsStore{}).
		withPaySession(r)

	payClient := (&mockPayClient{}).
		withASuccessfulPayment("abc123", "123456789012")

	dataStore := &mockDataStore{}
	dataStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(nil, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app", dataStore, mockRandom)(testAppData, w, r)

	assert.Equal(t, expectedError, errors.Unwrap(err))
	mock.AssertExpectationsForObjects(t, dataStore, lpaStore, sessionsStore, payClient)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r).
		withCompletedPaymentLpaData(r, "abc123", "123456789012")

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Email", mock.Anything, mock.Anything).
		Return("", nil)

	sessionsStore := (&mockSessionsStore{}).
		withPaySession(r)

	sessionsStore.
		On("Save", r, w, mock.Anything).
		Return(expectedError)

	logger := &mockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to expire cookie in session: %s", expectedError.Error())).
		Return(nil)

	payClient := (&mockPayClient{}).
		withASuccessfulPayment("abc123", "123456789012")

	dataStore := &mockDataStore{}
	dataStore.
		On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app", dataStore, mockRandom)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, lpaStore, sessionsStore, logger, payClient)
}

func (m *mockPayClient) withASuccessfulPayment(paymentId, reference string) *mockPayClient {
	m.
		On("GetPayment", paymentId).
		Return(pay.GetPaymentResponse{
			State: pay.State{
				Status:   "success",
				Finished: true,
			},
			PaymentId: paymentId,
			Reference: reference,
		}, nil)

	return m
}
