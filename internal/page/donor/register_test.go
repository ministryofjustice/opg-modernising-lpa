package donor

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
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
	Register(mux, &log.Logger{}, template.Templates{}, nil, nil, &onelogin.Client{}, &place.Client{}, "http://public.url", &pay.Client{}, nil, &mockWitnessCodeSender{}, nil, nil, nil, nil, &notify.Client{}, nil, nil, nil, nil)

	assert.Implements(t, (*http.Handler)(nil), mux)
}

func TestMakeHandle(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, nil)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			CanGoBack: true,
			SessionID: "cmFuZG9t",
			ActorType: actor.TypeDonor,
		}, appData)
		assert.Equal(t, w, hw)

		sessionData, _ := page.SessionDataFromContext(hr.Context())

		assert.Equal(t, &page.SessionData{SessionID: "cmFuZG9t"}, sessionData)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeHandleExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, nil)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
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

func TestMakeHandleErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, errorHandler.Execute)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
}

func TestMakeHandleSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, expectedError)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{}}, nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, sessionStore, None, nil)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error { return nil })

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeHandleNoSessionRequired(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()
	handle := makeHandle(mux, nil, None, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request) error {
		assert.Equal(t, page.AppData{
			Page:      "/path",
			ActorType: actor.TypeDonor,
		}, appData)
		assert.Equal(t, w, hw)
		assert.Equal(t, r.WithContext(page.ContextWithAppData(r.Context(), page.AppData{Page: "/path", ActorType: actor.TypeDonor})), hr)
		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeLpaHandleWhenDetailsProvidedAndUIDExists(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  actor.LpaTypePropertyFinance,
			Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted},
			UID:   "a-uid",
		}, nil)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, nil, donorStore)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		assert.Equal(t, page.AppData{
			Page:      "/lpa//path",
			ActorType: actor.TypeDonor,
			SessionID: "cmFuZG9t",
		}, appData)

		assert.Equal(t, w, hw)

		sessionData, _ := page.SessionDataFromContext(hr.Context())
		assert.Equal(t, &page.SessionData{SessionID: "cmFuZG9t"}, sessionData)

		hw.WriteHeader(http.StatusTeapot)
		return nil
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestMakeLpaHandleWhenSessionStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, expectedError)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, nil, nil)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestMakeLpaHandleWhenLpaStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	mux := http.NewServeMux()

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, expectedError)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	handle := makeLpaHandle(mux, sessionStore, RequireSession, errorHandler.Execute, donorStore)
	handle("/path", None, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
		return expectedError
	})

	mux.ServeHTTP(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMakeLpaHandleSessionExistingSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/path?a=b", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, None, nil, donorStore)
	handle("/path", RequireSession|CanGoBack, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
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
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)

	errorHandler := newMockErrorHandler(t)
	errorHandler.
		On("Execute", w, r, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{"session": &sesh.LoginSession{Sub: "random"}}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	mux := http.NewServeMux()
	handle := makeLpaHandle(mux, sessionStore, None, errorHandler.Execute, donorStore)
	handle("/path", RequireSession, func(appData page.AppData, hw http.ResponseWriter, hr *http.Request, lpa *page.Lpa) error {
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

			session := sessions.NewSession(sessionStore, "pay")
			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   5400,
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
				Secure:   true,
			}
			session.Values = map[any]any{"payment": &sesh.PaymentSession{PaymentID: "a-fake-id"}}

			sessionStore.
				On("Save", r, w, session).
				Return(nil)

			payClient := newMockPayClient(t)
			payClient.
				On("CreatePayment", pay.CreatePaymentBody{
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
				appPublicURL: "http://example.org",
				randomString: func(int) string { return "123456789012" },
			}).Pay(testAppData, w, r, &page.Lpa{ID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}, FeeType: pay.FullFee})
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
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID:               "lpa-id",
					FeeType:          feeType,
					Tasks:            actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
					EvidenceDelivery: pay.Upload,
				}).
				Return(nil)

			err := (&payHelper{
				donorStore: donorStore,
			}).Pay(testAppData, w, r, &page.Lpa{
				ID:               "lpa-id",
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
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID:               "lpa-id",
					FeeType:          feeType,
					Tasks:            actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
					EvidenceDelivery: pay.Post,
				}).
				Return(nil)

			err := (&payHelper{
				donorStore: donorStore,
			}).Pay(testAppData, w, r, &page.Lpa{
				ID:               "lpa-id",
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
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:               "lpa-id",
			FeeType:          pay.HalfFee,
			Tasks:            actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
			EvidenceDelivery: pay.Upload,
		}).
		Return(nil)

	err := (&payHelper{
		donorStore: donorStore,
	}).Pay(testAppData, w, r, &page.Lpa{
		ID:               "lpa-id",
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
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:      "lpa-id",
			FeeType: pay.NoFee,
			Tasks:   actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
		}).
		Return(expectedError)

	err := (&payHelper{
		donorStore: donorStore,
	}).Pay(testAppData, w, r, &page.Lpa{
		ID:      "lpa-id",
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

	session := sessions.NewSession(sessionStore, "pay")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   5400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{"payment": &sesh.PaymentSession{PaymentID: "a-fake-id"}}

	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	payClient := newMockPayClient(t)
	payClient.
		On("CreatePayment", pay.CreatePaymentBody{
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
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:             "lpa-id",
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
		appPublicURL: "http://example.org",
		randomString: func(int) string { return "123456789012" },
	}).Pay(testAppData, w, r, &page.Lpa{
		ID:             "lpa-id",
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

	session := sessions.NewSession(sessionStore, "pay")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   5400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{"payment": &sesh.PaymentSession{PaymentID: "a-fake-id"}}

	sessionStore.
		On("Save", r, w, session).
		Return(nil)

	payClient := newMockPayClient(t)
	payClient.
		On("CreatePayment", pay.CreatePaymentBody{
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
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:             "lpa-id",
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
		appPublicURL: "http://example.org",
		randomString: func(int) string { return "123456789012" },
	}).Pay(testAppData, w, r, &page.Lpa{
		ID:             "lpa-id",
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

	logger := newMockLogger(t)
	logger.
		On("Print", "Error creating payment: err")

	payClient := newMockPayClient(t)
	payClient.
		On("CreatePayment", mock.Anything).
		Return(pay.CreatePaymentResponse{}, expectedError)

	err := (&payHelper{
		logger:       logger,
		payClient:    payClient,
		appPublicURL: "http://example.org",
		randomString: func(int) string { return "123456789012" },
	}).Pay(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPayHelperPayWhenSessionErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Save", r, w, mock.Anything).
		Return(expectedError)

	payClient := newMockPayClient(t)
	payClient.
		On("CreatePayment", mock.Anything).
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
		appPublicURL: "http://example.org",
		randomString: func(int) string { return "123456789012" },
	}).Pay(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}
