package donorpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPay(t *testing.T) {
	testcases := map[string]struct {
		nextURL     string
		redirect    string
		canRedirect bool
	}{
		"real": {
			nextURL:     "https://www.payments.service.gov.uk/path-from/response",
			redirect:    "https://www.payments.service.gov.uk/path-from/response",
			canRedirect: true,
		},
		"fake": {
			nextURL:  "/lpa/lpa-id/something-else",
			redirect: page.Paths.PaymentConfirmation.Format("lpa-id"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				SetPayment(r, w, &sesh.PaymentSession{PaymentID: "a-fake-id"}).
				Return(nil)

			payClient := newMockPayClient(t)
			payClient.EXPECT().
				CreatePayment(r.Context(), "lpa-uid", pay.CreatePaymentBody{
					Amount:      8200,
					Reference:   "123456789012",
					Description: "Property and Finance LPA",
					ReturnURL:   "http://example.org/lpa/lpa-id/payment-confirmation",
					Email:       "a@b.com",
					Language:    "en",
				}).
				Return(&pay.CreatePaymentResponse{
					PaymentID: "a-fake-id",
					Links: map[string]pay.Link{
						"next_url": {
							Href: tc.nextURL,
						},
					},
				}, nil)
			payClient.EXPECT().
				CanRedirect(tc.nextURL).
				Return(tc.canRedirect)

			logger := newMockLogger(t)
			if !tc.canRedirect {
				logger.EXPECT().
					InfoContext(r.Context(), "skipping payment", slog.String("next_url", tc.nextURL))
			}

			err := Pay(logger, sessionStore, nil, payClient, func(int) string { return "123456789012" }, "http://example.org")(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id", LpaUID: "lpa-uid", Donor: donordata.Donor{Email: "a@b.com"}, FeeType: pay.FullFee})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPayWhenPaymentNotRequired(t *testing.T) {
	testCases := []pay.FeeType{
		pay.NoFee,
		pay.HardshipFee,
	}

	for _, feeType := range testCases {
		t.Run(feeType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.DonorProvidedDetails{
					LpaID:            "lpa-id",
					FeeType:          feeType,
					Tasks:            donordata.DonorTasks{PayForLpa: actor.PaymentTaskPending},
					EvidenceDelivery: pay.Upload,
				}).
				Return(nil)

			err := Pay(nil, nil, donorStore, nil, nil, "")(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaID:            "lpa-id",
				FeeType:          feeType,
				EvidenceDelivery: pay.Upload,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.EvidenceSuccessfullyUploaded.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPayWhenPostingEvidence(t *testing.T) {
	testCases := []pay.FeeType{
		pay.NoFee,
		pay.HardshipFee,
	}

	for _, feeType := range testCases {
		t.Run(feeType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.DonorProvidedDetails{
					LpaID:            "lpa-id",
					FeeType:          feeType,
					Tasks:            donordata.DonorTasks{PayForLpa: actor.PaymentTaskPending},
					EvidenceDelivery: pay.Post,
				}).
				Return(nil)

			err := Pay(nil, nil, donorStore, nil, nil, "")(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaID:            "lpa-id",
				FeeType:          feeType,
				EvidenceDelivery: pay.Post,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.WhatHappensNextPostEvidence.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPayWhenMoreEvidenceProvided(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID:            "lpa-id",
			FeeType:          pay.HalfFee,
			Tasks:            donordata.DonorTasks{PayForLpa: actor.PaymentTaskPending},
			EvidenceDelivery: pay.Upload,
		}).
		Return(nil)

	err := Pay(nil, nil, donorStore, nil, nil, "")(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID:            "lpa-id",
		FeeType:          pay.HalfFee,
		Tasks:            donordata.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired},
		EvidenceDelivery: pay.Upload,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.EvidenceSuccessfullyUploaded.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPayWhenPaymentNotRequiredWhenDonorStorePutError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID:   "lpa-id",
			FeeType: pay.NoFee,
			Tasks:   donordata.DonorTasks{PayForLpa: actor.PaymentTaskPending},
		}).
		Return(expectedError)

	err := Pay(nil, nil, donorStore, nil, nil, "")(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID:   "lpa-id",
		FeeType: pay.NoFee,
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPayWhenFeeDenied(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetPayment(r, w, &sesh.PaymentSession{PaymentID: "a-fake-id"}).
		Return(nil)

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		CreatePayment(r.Context(), "lpa-uid", pay.CreatePaymentBody{
			Amount:      4100,
			Reference:   "123456789012",
			Description: "Property and Finance LPA",
			ReturnURL:   "http://example.org/lpa/lpa-id/payment-confirmation",
			Email:       "a@b.com",
			Language:    "en",
		}).
		Return(&pay.CreatePaymentResponse{
			PaymentID: "a-fake-id",
			Links: map[string]pay.Link{
				"next_url": {
					Href: page.Paths.PaymentConfirmation.Format("lpa-id"),
				},
			},
		}, nil)
	payClient.EXPECT().
		CanRedirect(page.Paths.PaymentConfirmation.Format("lpa-id")).
		Return(false)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), mock.Anything, mock.Anything)

	err := Pay(logger, sessionStore, nil, payClient, func(int) string { return "123456789012" }, "http://example.org")(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID:          "lpa-id",
		LpaUID:         "lpa-uid",
		Donor:          donordata.Donor{Email: "a@b.com"},
		FeeType:        pay.HalfFee,
		Tasks:          donordata.DonorTasks{PayForLpa: actor.PaymentTaskDenied},
		PaymentDetails: []actor.Payment{{Amount: 4100}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id/payment-confirmation", resp.Header.Get("Location"))
}

func TestPayWhenCreatePaymentErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		CreatePayment(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := Pay(nil, nil, nil, payClient, func(int) string { return "123456789012" }, "")(testAppData, w, r, &donordata.DonorProvidedDetails{})

	assert.ErrorIs(t, err, expectedError)
}

func TestPayWhenSessionErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetPayment(r, w, mock.Anything).
		Return(expectedError)

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		CreatePayment(mock.Anything, mock.Anything, mock.Anything).
		Return(&pay.CreatePaymentResponse{
			PaymentID: "a-fake-id",
			Links: map[string]pay.Link{
				"next_url": {
					Href: "http://somewhere",
				},
			},
		}, nil)

	err := Pay(nil, sessionStore, nil, payClient, func(int) string { return "123456789012" }, "")(testAppData, w, r, &donordata.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}
