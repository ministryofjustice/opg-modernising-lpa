package donor

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
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

	template := &mockTemplate{}
	template.
		On("Func", w, &paymentConfirmationData{App: appData, PaymentReference: "123456789012", Continue: appData.Paths.TaskList}).
		Return(nil)

	sessionsStore := (&mockSessionsStore{}).
		withPaySession(r).
		withExpiredPaySession(r, w)

	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r).
		withCompletedPaymentLpaData(r, "abc123", "123456789012")

	err := PaymentConfirmation(&mockLogger{}, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, payClient, lpaStore, sessionsStore)
}

func TestGetPaymentConfirmationWhenDataStoreError(t *testing.T) {
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

	err := PaymentConfirmation(logger, template.Func, &mockPayClient{}, nil, lpaStore, &mockSessionsStore{}, "http://app")(appData, w, r)
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

	err := PaymentConfirmation(nil, template.Func, &mockPayClient{}, nil, lpaStore, sessionsStore, "http://app")(appData, w, r)
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

	err := PaymentConfirmation(logger, template.Func, payClient, nil, lpaStore, sessionsStore, "http://app")(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
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
		Return("", expectedError)

	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r)

	sessionsStore := (&mockSessionsStore{}).
		withPaySession(r)

	payClient := (&mockPayClient{}).
		withASuccessfulPayment("abc123", "123456789012")

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(nil, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(appData, w, r)

	assert.Equal(t, expectedError, errors.Unwrap(err))
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore, payClient)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Email", mock.Anything, mock.Anything).
		Return("", nil)

	lpaStore := (&mockLpaStore{}).
		willReturnEmptyLpa(r).
		withCompletedPaymentLpaData(r, "abc123", "123456789012")

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

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Func, payClient, notifyClient, lpaStore, sessionsStore, "http://app")(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionsStore, logger, payClient)
}

func (m *mockLpaStore) willReturnEmptyLpa(r *http.Request) *mockLpaStore {
	m.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
		}, nil)

	return m
}

func (m *mockLpaStore) withCompletedPaymentLpaData(r *http.Request, paymentId, paymentReference string) *mockLpaStore {
	m.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: page.PaymentDetails{
				PaymentId:        paymentId,
				PaymentReference: paymentReference,
			},
			Tasks: page.Tasks{
				PayForLpa: page.TaskCompleted,
			},
		}).
		Return(nil)

	return m
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

func (m *mockSessionsStore) withPaySession(r *http.Request) *mockSessionsStore {
	getSession := sessions.NewSession(m, "pay")

	getSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   5400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	getSession.Values = map[any]any{"payment": &sesh.PaymentSession{PaymentID: "abc123"}}

	m.On("Get", r, "pay").Return(getSession, nil)

	return m
}

func (m *mockSessionsStore) withExpiredPaySession(r *http.Request, w *httptest.ResponseRecorder) *mockSessionsStore {
	storeSession := sessions.NewSession(m, "pay")

	// Expire cookie
	storeSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	storeSession.Values = map[any]any{}
	m.On("Save", r, w, storeSession).Return(nil)

	return m
}
