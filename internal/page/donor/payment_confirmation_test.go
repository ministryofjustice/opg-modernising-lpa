package donor

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

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
	template.EXPECT().
		Execute(w, &paymentConfirmationData{
			App:              testAppData,
			PaymentReference: "123456789012",
			FeeType:          pay.FullFee,
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
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
	template.EXPECT().
		Execute(w, &paymentConfirmationData{
			App:              testAppData,
			PaymentReference: "123456789012",
			FeeType:          pay.HalfFee,
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
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
	sessionStore.EXPECT().
		Payment(r).
		Return(nil, expectedError)

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

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		GetPayment(r.Context(), "abc123").
		Return(pay.GetPaymentResponse{}, expectedError)

	template := newMockTemplate(t)

	err := PaymentConfirmation(nil, template.Execute, payClient, nil, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	donorStore := newMockDonorStore(t).
		withCompletedPaymentLpaData(r, "abc123", "123456789012", 8200)

	sessionStore := newMockSessionStore(t).
		withPaySession(r)
	sessionStore.EXPECT().
		ClearPayment(r, w).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "unable to expire cookie in session", slog.Any("err", expectedError))

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
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
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := PaymentConfirmation(nil, nil, payClient, donorStore, sessionStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		FeeType: pay.HalfFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
	})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func (m *mockPayClient) withASuccessfulPayment(paymentId, reference string, amount int, ctx context.Context) *mockPayClient {
	m.EXPECT().
		GetPayment(ctx, paymentId).
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
