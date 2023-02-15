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

	template := &page.MockTemplate{}
	template.
		On("Func", w, &paymentConfirmationData{App: page.TestAppData, PaymentReference: "123456789012", Continue: page.TestAppData.Paths.TaskList}).
		Return(nil)

	sessionsStore := (&page.MockSessionsStore{}).
		WithPaySession(r).
		WithExpiredPaySession(r, w)

	lpaStore := (&page.MockLpaStore{}).
		WillReturnEmptyLpa(r).
		WithCompletedPaymentLpaData(r, "abc123", "123456789012")

	err := PaymentConfirmation(&page.MockLogger{}, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, payClient, lpaStore, sessionsStore)
}

func TestGetPaymentConfirmationWhenDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := &page.MockTemplate{}
	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	logger := &page.MockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to retrieve item from data store using key '%s': %s", "session-id", page.ExpectedError.Error())).
		Return(nil)

	err := PaymentConfirmation(logger, template.Func, &mockPayClient{}, nil, lpaStore, &page.MockSessionsStore{}, "http://app")(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetPaymentConfirmationWhenErrorGettingSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := &page.MockTemplate{}
	lpaStore := (&page.MockLpaStore{}).
		WillReturnEmptyLpa(r)

	sessionsStore := &page.MockSessionsStore{}
	sessionsStore.
		On("Get", r, "pay").
		Return(&sessions.Session{}, page.ExpectedError)

	err := PaymentConfirmation(nil, template.Func, &mockPayClient{}, nil, lpaStore, sessionsStore, "http://app")(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore)
}

func TestGetPaymentConfirmationWhenErrorGettingPayment(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := (&page.MockLpaStore{}).
		WillReturnEmptyLpa(r)

	sessionsStore := (&page.MockSessionsStore{}).
		WithPaySession(r)

	logger := &page.MockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to retrieve payment info: %s", page.ExpectedError.Error())).
		Return(nil)

	payClient := &mockPayClient{}
	payClient.
		On("GetPayment", "abc123").
		Return(pay.GetPaymentResponse{}, page.ExpectedError)

	template := &page.MockTemplate{}

	err := PaymentConfirmation(logger, template.Func, payClient, nil, lpaStore, sessionsStore, "http://app")(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
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
		Return("", page.ExpectedError)

	lpaStore := (&page.MockLpaStore{}).
		WillReturnEmptyLpa(r)

	sessionsStore := (&page.MockSessionsStore{}).
		WithPaySession(r)

	payClient := (&mockPayClient{}).
		withASuccessfulPayment("abc123", "123456789012")

	template := &page.MockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(nil, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, errors.Unwrap(err))
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore, payClient)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := (&page.MockLpaStore{}).
		WillReturnEmptyLpa(r).
		WithCompletedPaymentLpaData(r, "abc123", "123456789012")

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Email", mock.Anything, mock.Anything).
		Return("", nil)

	sessionsStore := (&page.MockSessionsStore{}).
		WithPaySession(r)

	sessionsStore.
		On("Save", r, w, mock.Anything).
		Return(page.ExpectedError)

	logger := &page.MockLogger{}
	logger.
		On("Print", fmt.Sprintf("unable to expire cookie in session: %s", page.ExpectedError.Error())).
		Return(nil)

	payClient := (&mockPayClient{}).
		withASuccessfulPayment("abc123", "123456789012")

	template := &page.MockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(page.TestAppData, w, r)
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
