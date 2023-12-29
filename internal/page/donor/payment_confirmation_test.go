package donor

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentConfirmationFullFee(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

	template := newMockTemplate(t)
	template.
		On("Execute", w, &paymentConfirmationData{
			App:              testAppData,
			PaymentReference: "123456789012",
			FeeType:          pay.FullFee,
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.DonorProvidedDetails{
			FeeType: pay.FullFee,
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []actor.Payment{{
				PaymentId:        "abc123",
				PaymentReference: "123456789012",
				Amount:           8200,
			}},
			Tasks: actor.DonorTasks{
				PayForLpa: actor.PaymentTaskCompleted,
			},
		}).
		Return(nil)

	err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		FeeType: pay.FullFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationHalfFee(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 4100, r.Context())

	template := newMockTemplate(t)
	template.
		On("Execute", w, &paymentConfirmationData{
			App:              testAppData,
			PaymentReference: "123456789012",
			FeeType:          pay.HalfFee,
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.DonorProvidedDetails{
			FeeType: pay.HalfFee,
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []actor.Payment{{
				PaymentId:        "abc123",
				PaymentReference: "123456789012",
				Amount:           4100,
			}},
			Tasks: actor.DonorTasks{
				PayForLpa: actor.PaymentTaskPending,
			},
		}).
		Return(nil)

	err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		FeeType: pay.HalfFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorGettingSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := newMockTemplate(t)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "pay").
		Return(&sessions.Session{}, expectedError)

	err := PaymentConfirmation(nil, template.Execute, newMockPayClient(t), nil, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
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
		On("GetPayment", r.Context(), "abc123").
		Return(pay.GetPaymentResponse{}, expectedError)

	template := newMockTemplate(t)

	err := PaymentConfirmation(logger, template.Execute, payClient, nil, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	donorStore := newMockDonorStore(t).
		withCompletedPaymentLpaData(r, "abc123", "123456789012", 8200)

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
		withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Execute, payClient, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{CertificateProvider: actor.CertificateProvider{
		Email: "certificateprovider@example.com",
	}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationHalfFeeWhenDonorStorePutError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 4100, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", fmt.Sprintf("unable to update lpa in donorStore: %s", expectedError))

	err := PaymentConfirmation(logger, nil, payClient, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		FeeType: pay.HalfFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (m *mockPayClient) withASuccessfulPayment(paymentId, reference string, amount int, ctx context.Context) *mockPayClient {
	m.
		On("GetPayment", ctx, paymentId).
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
