package donor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"

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

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012")

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("Send", r.Context(), notify.CertificateProviderInviteEmail, testAppData, true, &page.Lpa{CertificateProvider: actor.CertificateProvider{Email: "certificateprovider@example.com"}}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &paymentConfirmationData{App: testAppData, PaymentReference: "123456789012"}).
		Return(nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	lpaStore := newMockLpaStore(t).
		willReturnEmptyLpa(r).
		withCompletedPaymentLpaData(r, "abc123", "123456789012")

	err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, lpaStore, sessionStore, shareCodeSender)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationGettingLpaErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := PaymentConfirmation(nil, nil, nil, lpaStore, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorGettingSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := newMockTemplate(t)
	lpaStore := newMockLpaStore(t).
		willReturnEmptyLpa(r)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "pay").
		Return(&sessions.Session{}, expectedError)

	err := PaymentConfirmation(nil, template.Execute, newMockPayClient(t), lpaStore, sessionStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorGettingPayment(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := newMockLpaStore(t).
		willReturnEmptyLpa(r)

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	logger := newMockLogger(t)
	logger.
		On("Print", fmt.Sprintf("unable to retrieve payment info: %s", expectedError.Error())).
		Return(nil)

	payClient := newMockPayClient(t)
	payClient.
		On("GetPayment", "abc123").
		Return(pay.GetPaymentResponse{}, expectedError)

	template := newMockTemplate(t)

	err := PaymentConfirmation(logger, template.Execute, payClient, lpaStore, sessionStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorSendingShareCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := newMockLpaStore(t).
		willReturnEmptyLpa(r)

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012")

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := PaymentConfirmation(nil, nil, payClient, lpaStore, sessionStore, shareCodeSender)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	lpaStore := newMockLpaStore(t).
		willReturnEmptyLpa(r).
		withCompletedPaymentLpaData(r, "abc123", "123456789012")

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	sessionStore.
		On("Save", r, w, mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", fmt.Sprintf("unable to expire cookie in session: %s", expectedError.Error())).
		Return(nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012")

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Execute, payClient, lpaStore, sessionStore, shareCodeSender)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
