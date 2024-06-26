package donor

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentConfirmationFullFee(t *testing.T) {
	testcases := map[string]struct {
		evidenceDelivery pay.EvidenceDelivery
		nextPage         page.LpaPath
	}{
		"empty": {
			nextPage: page.Paths.TaskList,
		},
		"upload": {
			evidenceDelivery: pay.Upload,
			nextPage:         page.Paths.EvidenceSuccessfullyUploaded,
		},
		"post": {
			evidenceDelivery: pay.Post,
			nextPage:         page.Paths.WhatHappensNextPostEvidence,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
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
					NextPage:         tc.nextPage,
					EvidenceDelivery: tc.evidenceDelivery,
				}).
				Return(nil)

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaUID:           "lpa-uid",
					FeeType:          pay.FullFee,
					EvidenceDelivery: tc.evidenceDelivery,
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

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendPaymentReceived(r.Context(), event.PaymentReceived{
					UID:       "lpa-uid",
					PaymentID: "abc123",
					Amount:    8200,
				}).
				Return(nil)

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaUID:           "lpa-uid",
				FeeType:          pay.FullFee,
				EvidenceDelivery: tc.evidenceDelivery,
				CertificateProvider: actor.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: actor.DonorTasks{
					PayForLpa: actor.PaymentTaskInProgress,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
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
			NextPage:         page.Paths.TaskList,
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaUID:  "lpa-uid",
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

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), event.PaymentReceived{
			UID:       "lpa-uid",
			PaymentID: "abc123",
			Amount:    4100,
		}).
		Return(nil)

	err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaUID:  "lpa-uid",
		FeeType: pay.HalfFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: actor.DonorTasks{
			PayForLpa: actor.PaymentTaskInProgress,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationApprovedOrDenied(t *testing.T) {
	for _, task := range []actor.PaymentTask{actor.PaymentTaskApproved, actor.PaymentTaskDenied} {
		t.Run(task.String(), func(t *testing.T) {
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
					NextPage:         page.Paths.TaskList,
				}).
				Return(nil)

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaUID:  "lpa-uid",
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

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendPaymentReceived(r.Context(), event.PaymentReceived{
					UID:       "lpa-uid",
					PaymentID: "abc123",
					Amount:    8200,
				}).
				Return(nil)

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: actor.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: actor.DonorTasks{
					PayForLpa: task,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPaymentConfirmationApprovedOrDeniedWhenSigned(t *testing.T) {
	for _, task := range []actor.PaymentTask{actor.PaymentTaskApproved, actor.PaymentTaskDenied} {
		t.Run(task.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			updatedDonor := &actor.DonorProvidedDetails{
				LpaUID:  "lpa-uid",
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
					PayForLpa:                  actor.PaymentTaskCompleted,
					ConfirmYourIdentityAndSign: actor.IdentityTaskCompleted,
				},
			}

			payClient := newMockPayClient(t).
				withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &paymentConfirmationData{
					App:              testAppData,
					PaymentReference: "123456789012",
					FeeType:          pay.FullFee,
					NextPage:         page.Paths.TaskList,
				}).
				Return(nil)

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), updatedDonor).
				Return(nil)

			shareCodeSender := newMockShareCodeSender(t)
			shareCodeSender.EXPECT().
				SendCertificateProviderPrompt(r.Context(), testAppData, updatedDonor).
				Return(nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendLpa(r.Context(), updatedDonor).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendPaymentReceived(r.Context(), event.PaymentReceived{
					UID:       "lpa-uid",
					PaymentID: "abc123",
					Amount:    8200,
				}).
				Return(nil)

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, shareCodeSender, lpaStoreClient, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: actor.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: actor.DonorTasks{
					PayForLpa:                  task,
					ConfirmYourIdentityAndSign: actor.IdentityTaskCompleted,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPaymentConfirmationApprovedOrDeniedWhenVoucherAllowed(t *testing.T) {
	for _, task := range []actor.PaymentTask{actor.PaymentTaskApproved, actor.PaymentTaskDenied} {
		t.Run(task.String(), func(t *testing.T) {
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
					NextPage:         page.Paths.WeHaveContactedVoucher,
				}).
				Return(nil)

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), mock.Anything).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendPaymentReceived(r.Context(), mock.Anything).
				Return(nil)

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: actor.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Voucher: actor.Voucher{Allowed: true},
				Tasks: actor.DonorTasks{
					PayForLpa: task,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPaymentConfirmationWhenNotSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		GetPayment(r.Context(), "abc123").
		Return(pay.GetPaymentResponse{
			State: pay.State{
				Status:   "error",
				Finished: true,
			},
		}, nil)

	err := PaymentConfirmation(newMockLogger(t), nil, payClient, nil, sessionStore, nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaUID: "lpa-uid",
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: actor.DonorTasks{
			PayForLpa: actor.PaymentTaskInProgress,
		},
	})

	assert.Error(t, err)
}

func TestGetPaymentConfirmationWhenErrorGettingSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	template := newMockTemplate(t)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Payment(r).
		Return(nil, expectedError)

	err := PaymentConfirmation(nil, template.Execute, newMockPayClient(t), nil, sessionStore, nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
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

	err := PaymentConfirmation(nil, template.Execute, payClient, nil, sessionStore, nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenErrorExpiringSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

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

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	err := PaymentConfirmation(logger, template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{CertificateProvider: actor.CertificateProvider{
		Email: "certificateprovider@example.com",
	}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenEventClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 4100, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(expectedError)

	err := PaymentConfirmation(nil, nil, payClient, nil, sessionStore, nil, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
		FeeType: pay.HalfFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
	})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationHalfFeeWhenDonorStorePutError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 4100, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	err := PaymentConfirmation(nil, nil, payClient, donorStore, sessionStore, nil, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
		FeeType: pay.HalfFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
	})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenLpaStoreClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(r.Context(), testAppData, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(r.Context(), mock.Anything).
		Return(expectedError)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	err := PaymentConfirmation(newMockLogger(t), nil, payClient, nil, sessionStore, shareCodeSender, lpaStoreClient, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
		FeeType: pay.FullFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: actor.DonorTasks{
			PayForLpa:                  actor.PaymentTaskApproved,
			ConfirmYourIdentityAndSign: actor.IdentityTaskCompleted,
		},
	})

	assert.ErrorIs(t, err, expectedError)
}

func TestGetPaymentConfirmationWhenShareCodeSenderErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(r.Context(), testAppData, mock.Anything).
		Return(expectedError)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	err := PaymentConfirmation(newMockLogger(t), nil, payClient, nil, sessionStore, shareCodeSender, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
		FeeType: pay.FullFee,
		CertificateProvider: actor.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: actor.DonorTasks{
			PayForLpa:                  actor.PaymentTaskApproved,
			ConfirmYourIdentityAndSign: actor.IdentityTaskCompleted,
		},
	})

	assert.ErrorIs(t, err, expectedError)
}

func (m *mockPayClient) withASuccessfulPayment(paymentId, reference string, amount int, ctx context.Context) *mockPayClient {
	m.EXPECT().
		GetPayment(ctx, paymentId).
		Return(pay.GetPaymentResponse{
			State: pay.State{
				Status:   "success",
				Finished: true,
			},
			PaymentID: paymentId,
			Reference: reference,
			Amount:    amount,
		}, nil)

	return m
}
