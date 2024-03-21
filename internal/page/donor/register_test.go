package donor

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, &slog.Logger{}, template.Templates{}, template.Templates{}, &mockSessionStore{}, &mockDonorStore{}, &onelogin.Client{}, &place.Client{}, "http://example.org", &pay.Client{}, &mockShareCodeSender{}, &mockWitnessCodeSender{}, nil, &mockCertificateProviderStore{}, &mockAttorneyStore{}, &mockNotifyClient{}, &mockEvidenceReceivedStore{}, &mockDocumentStore{}, &mockEventClient{}, &mockDashboardStore{}, &mockLpaStoreClient{}, &mockShareCodeStore{}, &mockProgressTracker{})

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, nil)
	handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			ActorType: actor.TypeDonor,
		}, appData)
		assert.Equal(t, w, hw)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleWhenRequireSession(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	loginSession := &sesh.LoginSession{Sub: "5", Email: "email"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(loginSession, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", page.RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:              "/path",
			ActorType:         actor.TypeDonor,
			SessionID:         loginSession.SessionID(),
			LoginSessionEmail: "email",
		}, appData)
		assert.Equal(t, w, hw)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleWhenRequireSessionErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, nil)
	handle("/path", page.RequireSession, nil)

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, errorHandler.Execute)
	handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeLpaHandleWhenDetailsProvidedAndUIDExists(t *testing.T) {
	testCases := map[string]struct {
		expectedAppData     page.AppData
		loginSesh           *sesh.LoginSession
		expectedSessionData *page.SessionData
	}{
		"donor": {
			expectedAppData: page.AppData{
				Page:      "/lpa/123/path",
				ActorType: actor.TypeDonor,
				SessionID: "cmFuZG9t",
				LpaID:     "123",
			},
			loginSesh:           &sesh.LoginSession{Sub: "random"},
			expectedSessionData: &page.SessionData{SessionID: "cmFuZG9t", LpaID: "123"},
		},
		"organisation": {
			expectedAppData: page.AppData{
				Page:        "/lpa/123/path",
				ActorType:   actor.TypeDonor,
				SessionID:   "cmFuZG9t",
				LpaID:       "123",
				IsSupporter: true,
				SupporterData: page.SupporterData{
					IsDonorPage:   true,
					DonorFullName: "Jane Smith",
					LpaType:       actor.LpaTypePropertyAndAffairs,
				},
			},
			loginSesh:           &sesh.LoginSession{Sub: "random", OrganisationID: "org-id"},
			expectedSessionData: &page.SessionData{SessionID: "cmFuZG9t", OrganisationID: "org-id", LpaID: "123"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/lpa/123/path", nil)

			mux := http.NewServeMux()

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Login(r).
				Return(tc.loginSesh, nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Get(mock.Anything).
				Return(&actor.DonorProvidedDetails{Donor: actor.Donor{
					FirstNames:  "Jane",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "1", "2"),
					Address:     place.Address{Postcode: "ABC123"},
				},
					Type:   actor.LpaTypePropertyAndAffairs,
					Tasks:  actor.DonorTasks{YourDetails: actor.TaskCompleted},
					LpaUID: "a-uid",
				}, nil)

			handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
			handle("/path", page.None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, _ *actor.DonorProvidedDetails) error {
				assert.Equal(t, tc.expectedAppData, appData)

				assert.Equal(t, w, hw)

				sessionData, _ := page.SessionDataFromContext(hr.Context())
				assert.Equal(t, tc.expectedSessionData, sessionData)

				hw.WriteHeader(http.StatusTeapot)
				return nil
			})

			mux.ServeHTTP(w, r)
			resp := w.Result()

			assert.Equal(t, http.StatusTeapot, resp.StatusCode)
		})
	}

}

func TestMakeLpaHandleWhenSessionStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/id/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	handle := makeLpaHandle(mux, sessionStore, nil, nil)
	handle("/path", page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.DonorProvidedDetails) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeLpaHandleWhenLpaStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/id/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.DonorProvidedDetails{}, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	handle := makeLpaHandle(mux, sessionStore, errorHandler.Execute, donorStore)
	handle("/path", page.None, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.DonorProvidedDetails) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeLpaHandleSessionExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "ignored"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/lpa/123/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.DonorProvidedDetails{}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, nil, donorStore)
	handle("/path", page.RequireSession|page.CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, _ *actor.DonorProvidedDetails) error {
		assert.Equal(t, page.AppData{
			Page:      "/lpa/123/path",
			SessionID: "cmFuZG9t",
			CanGoBack: true,
			LpaID:     "123",
			ActorType: actor.TypeDonor,
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := page.SessionDataFromContext(hr.Context())

		assert.Equal(t, &page.SessionData{LpaID: "123", SessionID: "cmFuZG9t"}, sessionData)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeLpaHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/lpa/123/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.EXPECT().
		Execute(w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "random"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.DonorProvidedDetails{}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, errorHandler.Execute, donorStore)
	handle("/path", page.RequireSession, func(_ page.AppData, _ http.ResponseWriter, _ *http.Request, _ *actor.DonorProvidedDetails) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestPayHelperPay(t *testing.T) {
	testcases := map[string]struct {
		nextURL  string
		redirect string
	}{
		"real": {
			nextURL:  "https://www.payments.service.gov.uk/path-from/response",
			redirect: "https://www.payments.service.gov.uk/path-from/response",
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
				CreatePayment(r.Context(), pay.CreatePaymentBody{
					Amount:      8200,
					Reference:   "123456789012",
					Description: "Property and Finance LPA",
					ReturnUrl:   "http://example.org/lpa/lpa-id/payment-confirmation",
					Email:       "a@b.com",
					Language:    "en",
				}).
				Return(pay.CreatePaymentResponse{
					PaymentId: "a-fake-id",
					Links: map[string]pay.Link{
						"next_url": {
							Href: tc.nextURL,
						},
					},
				}, nil)

			err := (&payHelper{
				sessionStore: sessionStore,
				payClient:    payClient,
				randomString: func(int) string { return "123456789012" },
				appPublicURL: "http://example.org",
			}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}, FeeType: pay.FullFee})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPayHelperPayWhenPaymentNotRequired(t *testing.T) {
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
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:            "lpa-id",
					FeeType:          feeType,
					Tasks:            actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
					EvidenceDelivery: pay.Upload,
				}).
				Return(nil)

			err := (&payHelper{
				donorStore: donorStore,
			}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{
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

func TestPayHelperPayWhenPostingEvidence(t *testing.T) {
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
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:            "lpa-id",
					FeeType:          feeType,
					Tasks:            actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
					EvidenceDelivery: pay.Post,
				}).
				Return(nil)

			err := (&payHelper{
				donorStore: donorStore,
			}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{
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

func TestPayHelperPayWhenMoreEvidenceProvided(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:            "lpa-id",
			FeeType:          pay.HalfFee,
			Tasks:            actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
			EvidenceDelivery: pay.Upload,
		}).
		Return(nil)

	err := (&payHelper{
		donorStore: donorStore,
	}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:            "lpa-id",
		FeeType:          pay.HalfFee,
		Tasks:            actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired},
		EvidenceDelivery: pay.Upload,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.EvidenceSuccessfullyUploaded.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPayHelperPayWhenPaymentNotRequiredWhenDonorStorePutError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:   "lpa-id",
			FeeType: pay.NoFee,
			Tasks:   actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
		}).
		Return(expectedError)

	err := (&payHelper{
		donorStore: donorStore,
	}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:   "lpa-id",
		FeeType: pay.NoFee,
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPayHelperPayWhenFeeDenied(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetPayment(r, w, &sesh.PaymentSession{PaymentID: "a-fake-id"}).
		Return(nil)

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		CreatePayment(r.Context(), pay.CreatePaymentBody{
			Amount:      4100,
			Reference:   "123456789012",
			Description: "Property and Finance LPA",
			ReturnUrl:   "http://example.org/lpa/lpa-id/payment-confirmation",
			Email:       "a@b.com",
			Language:    "en",
		}).
		Return(pay.CreatePaymentResponse{
			PaymentId: "a-fake-id",
			Links: map[string]pay.Link{
				"next_url": {
					Href: page.Paths.PaymentConfirmation.Format("lpa-id"),
				},
			},
		}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:          "lpa-id",
			Donor:          actor.Donor{Email: "a@b.com"},
			FeeType:        pay.FullFee,
			Tasks:          actor.DonorTasks{PayForLpa: actor.PaymentTaskInProgress},
			PaymentDetails: []actor.Payment{{Amount: 4100}},
		}).
		Return(nil)

	err := (&payHelper{
		sessionStore: sessionStore,
		donorStore:   donorStore,
		payClient:    payClient,
		randomString: func(int) string { return "123456789012" },
		appPublicURL: "http://example.org",
	}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:          "lpa-id",
		Donor:          actor.Donor{Email: "a@b.com"},
		FeeType:        pay.HalfFee,
		Tasks:          actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied},
		PaymentDetails: []actor.Payment{{Amount: 4100}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id/payment-confirmation", resp.Header.Get("Location"))
}

func TestPayHelperPayWhenFeeDeniedAndPutStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetPayment(r, w, &sesh.PaymentSession{PaymentID: "a-fake-id"}).
		Return(nil)

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		CreatePayment(r.Context(), pay.CreatePaymentBody{
			Amount:      4100,
			Reference:   "123456789012",
			Description: "Property and Finance LPA",
			ReturnUrl:   "http://example.org/lpa/lpa-id/payment-confirmation",
			Email:       "a@b.com",
			Language:    "en",
		}).
		Return(pay.CreatePaymentResponse{
			PaymentId: "a-fake-id",
			Links: map[string]pay.Link{
				"next_url": {
					Href: page.Paths.PaymentConfirmation.Format("lpa-id"),
				},
			},
		}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:          "lpa-id",
			Donor:          actor.Donor{Email: "a@b.com"},
			FeeType:        pay.FullFee,
			Tasks:          actor.DonorTasks{PayForLpa: actor.PaymentTaskInProgress},
			PaymentDetails: []actor.Payment{{Amount: 4100}},
		}).
		Return(expectedError)

	err := (&payHelper{
		sessionStore: sessionStore,
		donorStore:   donorStore,
		payClient:    payClient,
		randomString: func(int) string { return "123456789012" },
		appPublicURL: "http://example.org",
	}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:          "lpa-id",
		Donor:          actor.Donor{Email: "a@b.com"},
		FeeType:        pay.HalfFee,
		Tasks:          actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied},
		PaymentDetails: []actor.Payment{{Amount: 4100}},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPayHelperPayWhenCreatePaymentErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		CreatePayment(mock.Anything, mock.Anything).
		Return(pay.CreatePaymentResponse{}, expectedError)

	err := (&payHelper{
		payClient:    payClient,
		randomString: func(int) string { return "123456789012" },
	}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.ErrorIs(t, err, expectedError)
}

func TestPayHelperPayWhenSessionErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetPayment(r, w, mock.Anything).
		Return(expectedError)

	payClient := newMockPayClient(t)
	payClient.EXPECT().
		CreatePayment(mock.Anything, mock.Anything).
		Return(pay.CreatePaymentResponse{
			PaymentId: "a-fake-id",
			Links: map[string]pay.Link{
				"next_url": {
					Href: "http://somewhere",
				},
			},
		}, nil)

	err := (&payHelper{
		sessionStore: sessionStore,
		payClient:    payClient,
		randomString: func(int) string { return "123456789012" },
	}).Pay(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}
