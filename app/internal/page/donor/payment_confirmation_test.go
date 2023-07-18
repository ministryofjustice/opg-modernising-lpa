package donor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentConfirmation(t *testing.T) {
	testcases := map[page.FeeType]struct {
		State     actor.PaymentTask
		FeeAmount int
	}{
		page.FullFee:     {State: actor.PaymentTaskCompleted, FeeAmount: 8200},
		page.HalfFee:     {State: actor.PaymentTaskPending, FeeAmount: 4100},
		page.NoFee:       {State: actor.PaymentTaskPending, FeeAmount: 0},
		page.HardshipFee: {State: actor.PaymentTaskPending, FeeAmount: 0},
	}

	for fee, tc := range testcases {
		t.Run(fee.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			payClient := newMockPayClient(t).
				withASuccessfulPayment("abc123", "123456789012", tc.FeeAmount)

			shareCodeSender := newMockShareCodeSender(t)
			shareCodeSender.
				On("SendCertificateProvider", r.Context(), notify.CertificateProviderInviteEmail, testAppData, true, &page.Lpa{CertificateProvider: actor.CertificateProvider{Email: "certificateprovider@example.com"}, FeeType: fee}).
				Return(nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &paymentConfirmationData{App: testAppData, PaymentReference: "123456789012"}).
				Return(nil)

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					FeeType: fee,
					CertificateProvider: actor.CertificateProvider{
						Email: "certificateprovider@example.com",
					},
					PaymentDetails: page.PaymentDetails{
						PaymentId:        "abc123",
						PaymentReference: "123456789012",
						Amount:           tc.FeeAmount,
					},
					Tasks: page.Tasks{
						PayForLpa: tc.State,
					},
				}).
				Return(nil)

			reducedFeeStore := newMockReducedFeeStore(t)
			reducedFeeStore.
				On("Create", r.Context(), &page.Lpa{
					FeeType:             fee,
					CertificateProvider: actor.CertificateProvider{Email: "certificateprovider@example.com"},
					PaymentDetails: page.PaymentDetails{
						PaymentReference: "123456789012",
						PaymentId:        "abc123",
						Amount:           tc.FeeAmount,
					},
					Tasks: page.Tasks{PayForLpa: tc.State},
				}).
				Return(nil)

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, shareCodeSender, reducedFeeStore)(testAppData, w, r, &page.Lpa{
				FeeType: fee,
				CertificateProvider: actor.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPaymentConfirmationWhenErrorGettingSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := newMockTemplate(t)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "pay").
		Return(&sessions.Session{}, expectedError)

	err := PaymentConfirmation(nil, template.Execute, newMockPayClient(t), nil, sessionStore, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorGettingPayment(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

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

	err := PaymentConfirmation(logger, template.Execute, payClient, nil, sessionStore, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorSendingShareCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 8200)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := PaymentConfirmation(nil, nil, payClient, nil, sessionStore, shareCodeSender, nil)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{
		Email: "certificateprovider@example.com",
	}})

	assert.Equal(t, expectedError, err)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	donorStore := newMockDonorStore(t).
		withCompletedPaymentLpaData(r, "abc123", "123456789012")

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
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
		withASuccessfulPayment("abc123", "123456789012", 8200)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(nil)

	reducedFeeStore := newMockReducedFeeStore(t)
	reducedFeeStore.
		On("Create", r.Context(), mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Execute, payClient, donorStore, sessionStore, shareCodeSender, reducedFeeStore)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{
		Email: "certificateprovider@example.com",
	}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorCreatingReducedFee(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 8200)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", r.Context(), notify.CertificateProviderInviteEmail, testAppData, true, &page.Lpa{CertificateProvider: actor.CertificateProvider{Email: "certificateprovider@example.com"}}).
		Return(nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	donorStore := newMockDonorStore(t).
		withCompletedPaymentLpaData(r, "abc123", "123456789012")

	reducedFeeStore := newMockReducedFeeStore(t)
	reducedFeeStore.
		On("Create", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{Email: "certificateprovider@example.com"},
			PaymentDetails: page.PaymentDetails{
				PaymentReference: "123456789012",
				PaymentId:        "abc123",
				Amount:           8200,
			},
			Tasks: page.Tasks{PayForLpa: actor.PaymentTaskCompleted},
		}).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", "unable to create reduced fee: err").
		Return(nil)

	err := PaymentConfirmation(logger, nil, payClient, donorStore, sessionStore, shareCodeSender, reducedFeeStore)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{
		Email: "certificateprovider@example.com",
	}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (m *mockPayClient) withASuccessfulPayment(paymentId, reference string, amount int) *mockPayClient {
	m.
		On("GetPayment", paymentId).
		Return(pay.GetPaymentResponse{
			State: pay.State{
				Status:   "success",
				Finished: true,
			},
			PaymentId: paymentId,
			Reference: reference,
			Amount:    amount,
		}, nil)

	return m
}
