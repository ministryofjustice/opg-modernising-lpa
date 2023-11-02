package donor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentConfirmationFullFee(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 8200)

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
		On("Put", r.Context(), &page.Lpa{
			FeeType: pay.FullFee,
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []page.Payment{{
				PaymentId:        "abc123",
				PaymentReference: "123456789012",
				Amount:           8200,
			}},
			Tasks: page.Tasks{
				PayForLpa: actor.PaymentTaskCompleted,
			},
		}).
		Return(nil)

	err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil)(testAppData, w, r, &page.Lpa{
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
		withASuccessfulPayment("abc123", "123456789012", 4100)

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

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			FeeType: pay.HalfFee,
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []page.Payment{{
				PaymentId:        "abc123",
				PaymentReference: "123456789012",
				Amount:           4100,
			}},
			Tasks: page.Tasks{
				PayForLpa: actor.PaymentTaskPending,
			},
			Evidence: page.Evidence{Documents: []page.Document{
				{Key: "evidence-key", Sent: now},
				{Key: "another-evidence-key", Sent: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
			}},
		}).
		Return(nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObjectTagging", r.Context(), "evidence-key", []types.Tag{
			{Key: aws.String("replicate"), Value: aws.String("true")},
		}).
		Return(nil)

	err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, s3Client, func() time.Time { return now })(testAppData, w, r, &page.Lpa{
		FeeType: pay.HalfFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Evidence: page.Evidence{Documents: []page.Document{
			{Key: "evidence-key"},
			{Key: "another-evidence-key", Sent: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
		}},
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
		withASuccessfulPayment("abc123", "123456789012", 8200)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Execute, payClient, donorStore, sessionStore, nil, nil)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{
		Email: "certificateprovider@example.com",
	}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationHalfFeeWhenS3ClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 4100)

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	now := time.Now()

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObjectTagging", r.Context(), "evidence-key", []types.Tag{
			{Key: aws.String("replicate"), Value: aws.String("true")},
		}).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", fmt.Sprintf("error tagging evidence: %s", expectedError.Error())).
		Return(nil)

	err := PaymentConfirmation(logger, nil, payClient, nil, sessionStore, s3Client, func() time.Time { return now })(testAppData, w, r, &page.Lpa{
		FeeType: pay.HalfFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Evidence: page.Evidence{Documents: []page.Document{
			{Key: "evidence-key"},
		}},
	})
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
