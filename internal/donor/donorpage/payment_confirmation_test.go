package donorpage

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentConfirmationFullFee(t *testing.T) {
	testcases := map[string]struct {
		evidenceDelivery pay.EvidenceDelivery
		nextPage         donor.Path
	}{
		"empty": {
			nextPage: donor.PathTaskList,
		},
		"upload": {
			evidenceDelivery: pay.Upload,
			nextPage:         donor.PathEvidenceSuccessfullyUploaded,
		},
		"post": {
			evidenceDelivery: pay.Post,
			nextPage:         donor.PathPendingPayment,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			payClient := newMockPayClient(t).
				withASuccessfulPayment(8200, r.Context())

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					Type:             lpadata.LpaTypePersonalWelfare,
					Donor:            donordata.Donor{FirstNames: "a", LastName: "b"},
					LpaID:            "lpa-id",
					LpaUID:           "lpa-uid",
					FeeType:          pay.FullFee,
					EvidenceDelivery: tc.evidenceDelivery,
					CertificateProvider: donordata.CertificateProvider{
						Email: "certificateprovider@example.com",
					},
					PaymentDetails: []donordata.Payment{{
						PaymentID:        "abc123",
						PaymentReference: "123456789012",
						Amount:           8200,
					}},
					Tasks: donordata.Tasks{
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

			err := PaymentConfirmation(newMockLogger(t), payClient, donorStore, sessionStore, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				FeeType:          pay.FullFee,
				EvidenceDelivery: tc.evidenceDelivery,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStateInProgress,
				},
				Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
				Type:  lpadata.LpaTypePersonalWelfare,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathPaymentSuccessful.Format("lpa-id")+"?"+url.Values{
				"reference": {"123456789012"},
				"next":      {tc.nextPage.Format("lpa-id")},
			}.Encode(), resp.Header.Get("Location"))
		})
	}
}

func TestGetPaymentConfirmationHalfFee(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment(4100, r.Context())

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	sessionStore := newMockSessionStore(t).
		withPaySession(r).
		withExpiredPaySession(r, w)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			Type:    lpadata.LpaTypePersonalWelfare,
			Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
			LpaID:   "lpa-id",
			LpaUID:  "lpa-uid",
			FeeType: pay.HalfFee,
			CertificateProvider: donordata.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []donordata.Payment{{
				PaymentID:        "abc123",
				PaymentReference: "123456789012",
				Amount:           4100,
			}},
			Tasks: donordata.Tasks{
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

	err := PaymentConfirmation(newMockLogger(t), payClient, donorStore, sessionStore, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
		Type:    lpadata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		LpaID:   "lpa-id",
		LpaUID:  "lpa-uid",
		FeeType: pay.HalfFee,
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: donordata.Tasks{
			PayForLpa: task.PaymentStateInProgress,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathPaymentSuccessful.Format("lpa-id")+"?"+url.Values{
		"reference": {"123456789012"},
		"next":      {donor.PathTaskList.Format("lpa-id")},
	}.Encode(), resp.Header.Get("Location"))
}

func TestGetPaymentConfirmationRepeatApplicationFee(t *testing.T) {
	testcases := map[string]struct {
		evidenceDelivery pay.EvidenceDelivery
		nextPage         donor.Path
	}{
		"empty": {
			nextPage: donor.PathEvidenceSuccessfullyUploaded,
		},
		"upload": {
			evidenceDelivery: pay.Upload,
			nextPage:         donor.PathEvidenceSuccessfullyUploaded,
		},
		"post": {
			evidenceDelivery: pay.Post,
			nextPage:         donor.PathPendingPayment,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			payClient := newMockPayClient(t).
				withASuccessfulPayment(8200, r.Context())

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					Type:             lpadata.LpaTypePersonalWelfare,
					Donor:            donordata.Donor{FirstNames: "a", LastName: "b"},
					LpaID:            "lpa-id",
					LpaUID:           "lpa-uid",
					FeeType:          pay.RepeatApplicationFee,
					EvidenceDelivery: tc.evidenceDelivery,
					CertificateProvider: donordata.CertificateProvider{
						Email: "certificateprovider@example.com",
					},
					PaymentDetails: []donordata.Payment{{
						PaymentID:        "abc123",
						PaymentReference: "123456789012",
						Amount:           8200,
					}},
					Tasks: donordata.Tasks{
						PayForLpa: task.PaymentStatePending,
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

			err := PaymentConfirmation(newMockLogger(t), payClient, donorStore, sessionStore, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
				LpaUID:           "lpa-uid",
				LpaID:            "lpa-id",
				FeeType:          pay.RepeatApplicationFee,
				EvidenceDelivery: tc.evidenceDelivery,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStateInProgress,
				},
				Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
				Type:  lpadata.LpaTypePersonalWelfare,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathPaymentSuccessful.Format("lpa-id")+"?"+url.Values{
				"reference": {"123456789012"},
				"next":      {tc.nextPage.Format("lpa-id")},
			}.Encode(), resp.Header.Get("Location"))
		})
	}
}

func TestGetPaymentConfirmationApprovedOrDenied(t *testing.T) {
	for _, taskState := range []task.PaymentState{task.PaymentStateApproved, task.PaymentStateDenied} {
		t.Run(taskState.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			payClient := newMockPayClient(t).
				withASuccessfulPayment(8200, r.Context())

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					Type:    lpadata.LpaTypePersonalWelfare,
					Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
					LpaID:   "lpa-id",
					LpaUID:  "lpa-uid",
					FeeType: pay.FullFee,
					CertificateProvider: donordata.CertificateProvider{
						Email: "certificateprovider@example.com",
					},
					PaymentDetails: []donordata.Payment{{
						PaymentID:        "abc123",
						PaymentReference: "123456789012",
						Amount:           8200,
					}},
					Tasks: donordata.Tasks{
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

			err := PaymentConfirmation(newMockLogger(t), payClient, donorStore, sessionStore, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
				Type:    lpadata.LpaTypePersonalWelfare,
				Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaID:   "lpa-id",
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: donordata.Tasks{
					PayForLpa: taskState,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathPaymentSuccessful.Format("lpa-id")+"?"+url.Values{
				"reference": {"123456789012"},
				"next":      {donor.PathTaskList.Format("lpa-id")},
			}.Encode(), resp.Header.Get("Location"))
		})
	}
}

func TestGetPaymentConfirmationApprovedOrDeniedWhenSigned(t *testing.T) {
	for _, taskState := range []task.PaymentState{task.PaymentStateApproved, task.PaymentStateDenied} {
		t.Run(taskState.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			updatedDonor := &donordata.Provided{
				Type:    lpadata.LpaTypePersonalWelfare,
				Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaID:   "lpa-id",
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				PaymentDetails: []donordata.Payment{{
					PaymentID:        "abc123",
					PaymentReference: "123456789012",
					Amount:           8200,
				}},
				Tasks: donordata.Tasks{
					PayForLpa:  task.PaymentStateCompleted,
					SignTheLpa: task.StateCompleted,
				},
			}

			payClient := newMockPayClient(t).
				withASuccessfulPayment(8200, r.Context())

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

			sessionStore := newMockSessionStore(t).
				withPaySession(r).
				withExpiredPaySession(r, w)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), updatedDonor).
				Return(nil)

			accessCodeSender := newMockAccessCodeSender(t)
			accessCodeSender.EXPECT().
				SendCertificateProviderPrompt(r.Context(), testAppData, updatedDonor).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendPaymentReceived(r.Context(), event.PaymentReceived{
					UID:       "lpa-uid",
					PaymentID: "abc123",
					Amount:    8200,
				}).
				Return(nil)
			eventClient.EXPECT().
				SendCertificateProviderStarted(r.Context(), event.CertificateProviderStarted{
					UID: "lpa-uid",
				}).
				Return(nil)

			notifyClient := newMockNotifyClient(t).
				withEmailPersonalizations(r.Context(), "£82")

			err := PaymentConfirmation(newMockLogger(t), payClient, donorStore, sessionStore, accessCodeSender, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
				Type:    lpadata.LpaTypePersonalWelfare,
				Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaID:   "lpa-id",
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Tasks: donordata.Tasks{
					PayForLpa:  taskState,
					SignTheLpa: task.StateCompleted,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathPaymentSuccessful.Format("lpa-id")+"?"+url.Values{
				"reference": {"123456789012"},
				"next":      {donor.PathTaskList.Format("lpa-id")},
			}.Encode(), resp.Header.Get("Location"))
		})
	}
}

func TestGetPaymentConfirmationApprovedOrDeniedWhenVoucherAllowed(t *testing.T) {
	for _, task := range []task.PaymentState{task.PaymentStateApproved, task.PaymentStateDenied} {
		t.Run(task.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

			payClient := newMockPayClient(t).
				withASuccessfulPayment(8200, r.Context())

			localizer := newMockLocalizer(t).
				withEmailLocalizations()

			testAppData.Localizer = localizer

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

			provided := &donordata.Provided{
				Type:    lpadata.LpaTypePersonalWelfare,
				Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaID:   "lpa-id",
				LpaUID:  "lpa-uid",
				FeeType: pay.FullFee,
				CertificateProvider: donordata.CertificateProvider{
					Email: "certificateprovider@example.com",
				},
				Voucher: donordata.Voucher{Allowed: true},
				Tasks: donordata.Tasks{
					PayForLpa: task,
				},
			}

			accessCodeSender := newMockAccessCodeSender(t)
			accessCodeSender.EXPECT().
				SendVoucherInvite(r.Context(), provided, testAppData).
				Return(nil)

			err := PaymentConfirmation(newMockLogger(t), payClient, donorStore, sessionStore, accessCodeSender, eventClient, notifyClient)(testAppData, w, r, provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathPaymentSuccessful.Format("lpa-id")+"?"+url.Values{
				"reference": {"123456789012"},
				"next":      {donor.PathWeHaveContactedVoucher.Format("lpa-id")},
			}.Encode(), resp.Header.Get("Location"))
		})
	}
}

func TestGetPaymentConfirmationWhenVoucherAllowedAccessCodeError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment(8200, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t).
		withEmailPersonalizations(r.Context(), "£82")

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendVoucherInvite(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := PaymentConfirmation(newMockLogger(t), payClient, nil, sessionStore, accessCodeSender, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
		Type:    lpadata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		LpaUID:  "lpa-uid",
		FeeType: pay.FullFee,
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Voucher: donordata.Voucher{Allowed: true},
		Tasks: donordata.Tasks{
			PayForLpa: task.PaymentStateDenied,
		},
	})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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

	err := PaymentConfirmation(newMockLogger(t), payClient, nil, sessionStore, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaUID: "lpa-uid",
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: donordata.Tasks{
			PayForLpa: task.PaymentStateInProgress,
		},
	})

	assert.Error(t, err)
}

func TestGetPaymentConfirmationWhenErrorGettingSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Payment(r).
		Return(nil, expectedError)

	err := PaymentConfirmation(nil, newMockPayClient(t), nil, sessionStore, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
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

	err := PaymentConfirmation(nil, payClient, nil, sessionStore, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
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
		withASuccessfulPayment(8200, r.Context())

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t).
		withEmailPersonalizations(r.Context(), "£82")

	err := PaymentConfirmation(logger, payClient, donorStore, sessionStore, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Type:   lpadata.LpaTypePersonalWelfare,
		Donor:  donordata.Donor{FirstNames: "a", LastName: "b"},
		LpaUID: "lpa-uid",
		LpaID:  "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathPaymentSuccessful.Format("lpa-id")+"?"+url.Values{
		"reference": {"123456789012"},
		"next":      {donor.PathTaskList.Format("lpa-id")},
	}.Encode(), resp.Header.Get("Location"))
}

func TestGetPaymentConfirmationWhenEventClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment(4100, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(expectedError)

	err := PaymentConfirmation(nil, payClient, nil, sessionStore, nil, eventClient, nil)(testAppData, w, r, &donordata.Provided{
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
		withASuccessfulPayment(4100, r.Context())

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

	err := PaymentConfirmation(nil, payClient, nil, sessionStore, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
		Type:    lpadata.LpaTypePersonalWelfare,
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
		withASuccessfulPayment(4100, r.Context())

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

	err := PaymentConfirmation(nil, payClient, donorStore, sessionStore, nil, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
		LpaUID:  "lpa-uid",
		Type:    lpadata.LpaTypePersonalWelfare,
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

func TestGetPaymentConfirmationWhenEventClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment(8200, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendCertificateProviderPrompt(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPaymentReceived(r.Context(), mock.Anything).
		Return(nil)
	eventClient.EXPECT().
		SendCertificateProviderStarted(r.Context(), mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t).
		withEmailLocalizations()

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t).
		withEmailPersonalizations(r.Context(), "£82")

	err := PaymentConfirmation(newMockLogger(t), payClient, nil, sessionStore, accessCodeSender, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
		LpaUID:  "lpa-uid",
		Type:    lpadata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		FeeType: pay.FullFee,
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: donordata.Tasks{
			PayForLpa:  task.PaymentStateApproved,
			SignTheLpa: task.StateCompleted,
		},
	})

	assert.ErrorIs(t, err, expectedError)
}

func TestGetPaymentConfirmationWhenAccessCodeSenderErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

	payClient := newMockPayClient(t).
		withASuccessfulPayment(8200, r.Context())

	sessionStore := newMockSessionStore(t).
		withPaySession(r)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
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

	err := PaymentConfirmation(newMockLogger(t), payClient, nil, sessionStore, accessCodeSender, eventClient, notifyClient)(testAppData, w, r, &donordata.Provided{
		LpaUID:  "lpa-uid",
		Type:    lpadata.LpaTypePersonalWelfare,
		Donor:   donordata.Donor{FirstNames: "a", LastName: "b"},
		FeeType: pay.FullFee,
		CertificateProvider: donordata.CertificateProvider{
			Email: "certificateprovider@example.com",
		},
		Tasks: donordata.Tasks{
			PayForLpa:  task.PaymentStateApproved,
			SignTheLpa: task.StateCompleted,
		},
	})

	assert.ErrorIs(t, err, expectedError)
}

func (m *mockPayClient) withASuccessfulPayment(amount int, ctx context.Context) *mockPayClient {
	m.EXPECT().
		GetPayment(ctx, "abc123").
		Return(pay.GetPaymentResponse{
			Email: "a@example.com",
			State: pay.State{
				Status:   "success",
				Finished: true,
			},
			PaymentID:   "abc123",
			Reference:   "123456789012",
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
		Return("donor name possessive").
		Once()
	m.EXPECT().
		T("personal-welfare").
		Return("translated type").
		Once()
	m.EXPECT().
		FormatDate(time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)).
		Return("formatted capture submit time").
		Once()
	return m
}

func (m *mockNotifyClient) withEmailPersonalizations(ctx context.Context, amount string) *mockNotifyClient {
	m.EXPECT().
		SendEmail(ctx, notify.ToCustomEmail(localize.En, "a@example.com"), notify.PaymentConfirmationEmail{
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
