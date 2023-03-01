package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var publicUrl = "http://example.org"

func TestGetAboutPayment(t *testing.T) {
	random := func(int) string { return "123456789012" }

	t.Run("Handles page data", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Get", r.Context()).
			Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, nil)

		template := newMockTemplate(t)
		template.
			On("Execute", w, &aboutPaymentData{App: testAppData}).
			Return(nil)

		payClient := newMockPayClient(t)

		err := AboutPayment(nil, template.Execute, nil, payClient, publicUrl, random, lpaStore)(testAppData, w, r)
		resp := w.Result()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Returns error when an cannot return LPA from store", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Get", r.Context()).
			Return(&page.Lpa{}, expectedError)

		err := AboutPayment(nil, nil, nil, nil, publicUrl, random, lpaStore)(testAppData, w, r)
		resp := w.Result()

		assert.Equal(t, err, expectedError)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Returns an error when cannot render template", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Get", r.Context()).
			Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, nil)

		template := newMockTemplate(t)
		template.
			On("Execute", w, &aboutPaymentData{App: testAppData}).
			Return(expectedError)

		err := AboutPayment(nil, template.Execute, nil, nil, publicUrl, random, lpaStore)(testAppData, w, r)
		resp := w.Result()

		assert.Equal(t, expectedError, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestPostAboutPayment(t *testing.T) {
	random := func(int) string { return "123456789012" }

	t.Run("Creates GOV UK Pay payment and saves paymentID in secure cookie", func(t *testing.T) {
		testCases := map[string]struct {
			nextUrl  string
			redirect string
		}{
			"Real return URL": {
				nextUrl:  "https://www.payments.service.gov.uk/path-from/response",
				redirect: "https://www.payments.service.gov.uk/path-from/response",
			},
			"Fake return URL": {
				nextUrl:  "/lpa/lpa-id/something-else",
				redirect: "/lpa/lpa-id" + page.Paths.PaymentConfirmation,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

				lpaStore := newMockLpaStore(t)
				lpaStore.
					On("Get", r.Context()).
					Return(&page.Lpa{You: actor.Person{Email: "a@b.com"}, CertificateProvider: actor.CertificateProvider{}}, nil)

				template := newMockTemplate(t)
				template.
					On("Execute", w, &aboutPaymentData{App: testAppData}).
					Return(nil)

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
								Href: tc.nextUrl,
							},
						},
					}, nil)

				err := AboutPayment(nil, template.Execute, sessionStore, payClient, publicUrl, random, lpaStore)(testAppData, w, r)
				resp := w.Result()

				assert.Nil(t, err)
				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, tc.redirect, resp.Header.Get("Location"))

			})
		}
	})

	t.Run("Returns error when cannot create payment", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Get", r.Context()).
			Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, nil)

		template := newMockTemplate(t)

		sessionStore := newMockSessionStore(t)

		logger := newMockLogger(t)
		logger.
			On("Print", "Error creating payment: "+expectedError.Error())

		payClient := newMockPayClient(t)
		payClient.
			On("CreatePayment", mock.Anything).
			Return(pay.CreatePaymentResponse{}, expectedError)

		err := AboutPayment(logger, template.Execute, sessionStore, payClient, publicUrl, random, lpaStore)(testAppData, w, r)

		assert.Equal(t, expectedError, err, "Expected error was not returned")
	})

	t.Run("Returns error when cannot save to session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Get", r.Context()).
			Return(&page.Lpa{CertificateProvider: actor.CertificateProvider{}}, nil)

		template := newMockTemplate(t)

		sessionStore := newMockSessionStore(t)

		sessionStore.
			On("Save", mock.Anything, mock.Anything, mock.Anything).
			Return(expectedError)

		logger := newMockLogger(t)

		payClient := newMockPayClient(t)
		payClient.
			On("CreatePayment", mock.Anything).
			Return(pay.CreatePaymentResponse{Links: map[string]pay.Link{"next_url": {Href: "http://example.url"}}}, nil)

		err := AboutPayment(logger, template.Execute, sessionStore, payClient, publicUrl, random, lpaStore)(testAppData, w, r)

		assert.Equal(t, expectedError, err, "Expected error was not returned")
	})
}
