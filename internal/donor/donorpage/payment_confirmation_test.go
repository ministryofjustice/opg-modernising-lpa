package donorpage

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
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

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

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
				Put(r.Context(), &donordata.DonorProvidedDetails{
					Type:             donordata.LpaTypePersonalWelfare,
					Donor:            donordata.Donor{FirstNames: "a", LastName: "b"},
					LpaUID:           "lpa-uid",
					FeeType:          pay.FullFee,
					EvidenceDelivery: tc.evidenceDelivery,
					CertificateProvider: donordata.CertificateProvider{
						Email: "certificateprovider@example.com",
					},
					PaymentDetails: []donordata.Payment{{
						PaymentId:        "abc123",
						PaymentReference: "123456789012",
						Amount:           8200,
					}},
					Tasks: donordata.DonorTasks{
						PayForLpa: task.PaymentStateCompleted,
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

			notifyClient := newMockNotifyClient(t).
				withEmailPersonalizations(r.Context(), "£82")

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
				LpaUID:           "lpa-uid",
				FeeType:          pay.FullFee,
				EvidenceDelivery: tc.evidenceDelivery,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: donordata.DonorTasks{
					PayForLpa: task.PaymentStateInProgress,
				},
				Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
				Type:  donordata.LpaTypePersonalWelfare,
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

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

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
		Put(r.Context(), &donordata.DonorProvidedDetails{
			Type:    donordata.LpaTypePersonalWelfare,
			Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
			LpaUID:  "lpa-uid",
			FeeType: pay.HalfFee,
			CertificateProvider: donordata.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []donordata.Payment{{
				PaymentId:        "abc123",
				PaymentReference: "123456789012",
				Amount:           4100,
			}},
			Tasks: donordata.DonorTasks{
				PayForLpa: task.PaymentStatePending,
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

	notifyClient := newMockNotifyClient(t).
		withEmailPersonalizations(r.Context(), "£41")

	err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
		Type:    donordata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		LpaUID:  "lpa-uid",
		FeeType: pay.HalfFee,
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: donordata.DonorTasks{
			PayForLpa: task.PaymentStateInProgress,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationApprovedOrDenied(t *testing.T) {
	for _, taskState := range []task.PaymentState{task.PaymentStateApproved, task.PaymentStateDenied} {
		t.Run(taskState.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			payClient := newMockPayClient(t).
				withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

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
				Put(r.Context(), &donordata.DonorProvidedDetails{
					Type:    donordata.LpaTypePersonalWelfare,
					Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
					LpaUID:  "lpa-uid",
					FeeType: pay.FullFee,
					CertificateProvider: donordata.CertificateProvider{
						Email: "certificateprovider@example.com",
					},
					PaymentDetails: []donordata.Payment{{
						PaymentId:        "abc123",
						PaymentReference: "123456789012",
						Amount:           8200,
					}},
					Tasks: donordata.DonorTasks{
						PayForLpa: task.PaymentStateCompleted,
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

			notifyClient := newMockNotifyClient(t).
				withEmailPersonalizations(r.Context(), "£82")

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
				Type:    donordata.LpaTypePersonalWelfare,
				Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: donordata.DonorTasks{
					PayForLpa: taskState,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPaymentConfirmationApprovedOrDeniedWhenSigned(t *testing.T) {
	for _, taskState := range []task.PaymentState{task.PaymentStateApproved, task.PaymentStateDenied} {
		t.Run(taskState.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			updatedDonor := &donordata.DonorProvidedDetails{
				Type:    donordata.LpaTypePersonalWelfare,
				Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				PaymentDetails: []donordata.Payment{{
					PaymentId:        "abc123",
					PaymentReference: "123456789012",
					Amount:           8200,
				}},
				Tasks: donordata.DonorTasks{
					PayForLpa:                  task.PaymentStateCompleted,
					ConfirmYourIdentityAndSign: actor.IdentityTaskCompleted,
				},
			}

			payClient := newMockPayClient(t).
				withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

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

			notifyClient := newMockNotifyClient(t).
				withEmailPersonalizations(r.Context(), "£82")

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, shareCodeSender, lpaStoreClient, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
				Type:    donordata.LpaTypePersonalWelfare,
				Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: donordata.DonorTasks{
					PayForLpa:                  taskState,
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
	for _, task := range []task.PaymentState{task.PaymentStateApproved, task.PaymentStateDenied} {
		t.Run(task.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			payClient := newMockPayClient(t).
				withASuccessfulPayment("abc123", "123456789012", 8200, r.Context())

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

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

			notifyClient := newMockNotifyClient(t).
				withEmailPersonalizations(r.Context(), "£82")

			err := PaymentConfirmation(newMockLogger(t), template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
				Type:    donordata.LpaTypePersonalWelfare,
				Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Voucher: donordata.Voucher{Allowed: true},
				Tasks: donordata.DonorTasks{
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

	err := PaymentConfirmation(newMockLogger(t), nil, payClient, nil, sessionStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaUID: "lpa-uid",
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: donordata.DonorTasks{
			PayForLpa: task.PaymentStateInProgress,
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

	err := PaymentConfirmation(nil, template.Execute, newMockPayClient(t), nil, sessionStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
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

	err := PaymentConfirmation(nil, template.Execute, payClient, nil, sessionStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
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

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t).
		withEmailPersonalizations(r.Context(), "£82")

	err := PaymentConfirmation(logger, template.Execute, payClient, donorStore, sessionStore, nil, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Type:   donordata.LpaTypePersonalWelfare,
		Donor:  donordata.Donor{FirstNames: "a", LastName: "b"},
		LpaUID: "lpa-uid"})
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

	err := PaymentConfirmation(nil, nil, payClient, nil, sessionStore, nil, nil, eventClient, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		FeeType: pay.HalfFee,
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
	})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPaymentConfirmationWhenNotifyClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment("abc123", "123456789012", 4100, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendEmail(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	err := PaymentConfirmation(nil, nil, payClient, nil, sessionStore, nil, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
		Type:    donordata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		FeeType: pay.HalfFee,
		CertificateProvider: donordata.CertificateProvider{
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

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t).
		withEmailPersonalizations(r.Context(), "£41")

	err := PaymentConfirmation(nil, nil, payClient, donorStore, sessionStore, nil, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaUID:  "lpa-uid",
		Type:    donordata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		FeeType: pay.HalfFee,
		CertificateProvider: donordata.CertificateProvider{
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
		SendCertificateProviderPrompt(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(r.Context(), mock.Anything).
		Return(expectedError)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t).
		withEmailPersonalizations(r.Context(), "£82")

	err := PaymentConfirmation(newMockLogger(t), nil, payClient, nil, sessionStore, shareCodeSender, lpaStoreClient, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaUID:  "lpa-uid",
		Type:    donordata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		FeeType: pay.FullFee,
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: donordata.DonorTasks{
			PayForLpa:                  task.PaymentStateApproved,
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
		SendCertificateProviderPrompt(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t).
		withEmailPersonalizations(r.Context(), "£82")

	err := PaymentConfirmation(newMockLogger(t), nil, payClient, nil, sessionStore, shareCodeSender, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaUID:  "lpa-uid",
		Type:    donordata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		FeeType: pay.FullFee,
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: donordata.DonorTasks{
			PayForLpa:                  task.PaymentStateApproved,
			ConfirmYourIdentityAndSign: actor.IdentityTaskCompleted,
		},
	})

	assert.ErrorIs(t, err, expectedError)
}

func (m *mockPayClient) withASuccessfulPayment(paymentId, reference string, amount int, ctx context.Context) *mockPayClient {
	m.EXPECT().
		GetPayment(ctx, paymentId).
		Return(pay.GetPaymentResponse{
			Email: "a@example.com",
			State: pay.State{
				Status:   "success",
				Finished: true,
			},
			PaymentID:   paymentId,
			Reference:   reference,
			AmountPence: pay.AmountPence(amount),
			SettlementSummary: pay.SettlementSummary{
				CaptureSubmitTime: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC),
				CapturedDate:      date.New("2000", "01", "02"),
			},
			CardDetails: pay.CardDetails{CardholderName: "a b"},
		}, nil)

	return m
}

func (m *mockLocalizer) withEmailLocalizations() *mockLocalizer {
	m.EXPECT().
		Possessive("a b").
		Return("donor name possessive")
	m.EXPECT().
		T("personal-welfare").
		Return("translated type")
	m.EXPECT().
		FormatDate(time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)).
		Return("formatted capture submit time")
	return m
}

func (m *mockNotifyClient) withEmailPersonalizations(ctx context.Context, amount string) *mockNotifyClient {
	m.EXPECT().
		SendEmail(ctx, "a@example.com", notify.PaymentConfirmationEmail{
			DonorFullNamesPossessive: "donor name possessive",
			LpaType:                  "translated type",
			PaymentCardFullName:      "a b",
			LpaReferenceNumber:       "lpa-uid",
			PaymentReferenceID:       "abc123",
			PaymentConfirmationDate:  "formatted capture submit time",
			AmountPaidWithCurrency:   amount,
		}).
		Return(nil)
	return m
}
