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
				"link": fmt.Sprintf("http://app%s?lpaId=lpa-id&sessionId=session-id", page.Paths.CertificateProviderStart),
			},
		}).
		Return("", nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &paymentConfirmationData{App: TestAppData, PaymentReference: "123456789012", Continue: TestAppData.Paths.TaskList}).
		Return(nil)

	sessionsStore := (&MockSessionsStore{}).
		WithPaySession(r).
		WithExpiredPaySession(r, w)

	lpaStore := (&MockLpaStore{}).
		WillReturnEmptyLpa(r).
		WithCompletedPaymentLpaData(r, "abc123", "123456789012")

	err := PaymentConfirmation(&MockLogger{}, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, payClient, lpaStore, sessionsStore)
}

func TestGetPaymentConfirmationWhenDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := &MockTemplate{}
	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, ExpectedError)

	logger := &MockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to retrieve item from data store using key '%s': %s", "session-id", ExpectedError.Error())).
		Return(nil)

	err := PaymentConfirmation(logger, template.Func, &mockPayClient{}, nil, lpaStore, &MockSessionsStore{}, "http://app")(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetPaymentConfirmationWhenErrorGettingSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := &MockTemplate{}
	lpaStore := (&MockLpaStore{}).
		WillReturnEmptyLpa(r)

	sessionsStore := &MockSessionsStore{}
	sessionsStore.
		On("Get", r, "pay").
		Return(&sessions.Session{}, ExpectedError)

	err := PaymentConfirmation(nil, template.Func, &mockPayClient{}, nil, lpaStore, sessionsStore, "http://app")(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore)
}

func TestGetPaymentConfirmationWhenErrorGettingPayment(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := (&MockLpaStore{}).
		WillReturnEmptyLpa(r)

	sessionsStore := (&MockSessionsStore{}).
		WithPaySession(r)

	logger := &MockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to retrieve payment info: %s", ExpectedError.Error())).
		Return(nil)

	payClient := &mockPayClient{}
	payClient.
		On("GetPayment", "abc123").
		Return(pay.GetPaymentResponse{}, ExpectedError)

	template := &MockTemplate{}

	err := PaymentConfirmation(logger, template.Func, payClient, nil, lpaStore, sessionsStore, "http://app")(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore, logger, payClient)
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
		Return("", ExpectedError)

	lpaStore := (&MockLpaStore{}).
		WillReturnEmptyLpa(r)

	sessionsStore := (&MockSessionsStore{}).
		WithPaySession(r)

	payClient := (&mockPayClient{}).
		withASuccessfulPayment("abc123", "123456789012")

	template := &MockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(nil, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(TestAppData, w, r)

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore, payClient)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := (&MockLpaStore{}).
		WillReturnEmptyLpa(r).
		WithCompletedPaymentLpaData(r, "abc123", "123456789012")

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Email", mock.Anything, mock.Anything).
		Return("", nil)

	sessionsStore := (&MockSessionsStore{}).
		WithPaySession(r)

	sessionsStore.
		On("Save", r, w, mock.Anything).
		Return(ExpectedError)

	logger := &MockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to expire cookie in session: %s", ExpectedError.Error())).
		Return(nil)

	payClient := (&mockPayClient{}).
		withASuccessfulPayment("abc123", "123456789012")

	template := &MockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore, logger, payClient)
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
