package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"

	"github.com/gorilla/sessions"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPayClient struct {
	mock.Mock
	BaseURL string
}

func (m *mockPayClient) CreatePayment(body pay.CreatePaymentBody) (pay.CreatePaymentResponse, error) {
	args := m.Called(body)
	return args.Get(0).(pay.CreatePaymentResponse), args.Error(1)
}

func (m *mockPayClient) GetPayment(paymentId string) (pay.GetPaymentResponse, error) {
	args := m.Called(paymentId)
	return args.Get(0).(pay.GetPaymentResponse), args.Error(1)
}

func TestAboutPayment(t *testing.T) {
	payClient := mockPayClient{BaseURL: "http://base.url"}

	t.Run("GET", func(t *testing.T) {
		t.Run("Handles page data", func(t *testing.T) {
			w := httptest.NewRecorder()
			appData := AppData{}

			template := &mockTemplate{}
			template.
				On("Func", w, &aboutPaymentData{App: appData}).
				Return(nil)

			r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

			err := AboutPayment(&mockLogger{}, template.Func, &mockSessionsStore{}, &payClient)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, template)
		})

		t.Run("Returns an error when cannot render template", func(t *testing.T) {
			w := httptest.NewRecorder()
			appData := AppData{}

			template := &mockTemplate{}
			template.
				On("Func", w, &aboutPaymentData{App: appData}).
				Return(expectedError)

			r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

			err := AboutPayment(&mockLogger{}, template.Func, &mockSessionsStore{}, &payClient)(appData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, template)
		})
	})

	t.Run("POST", func(t *testing.T) {
		t.Run("Creates GOV UK Pay payment and saves paymentID in secure cookie", func(t *testing.T) {
			testCases := map[string]struct {
				baseUrl                 string
				expectedRedirectPath    string
				expectCookieSecureValue bool
			}{
				"Real base URL": {
					baseUrl:                 "https://publicapi.payments.service.gov.uk",
					expectedRedirectPath:    "https://publicapi.payments.service.gov.uk/path-from/response",
					expectCookieSecureValue: true,
				},
				"Mock base URL": {
					baseUrl:                 "http://mock-pay.com",
					expectedRedirectPath:    "/payment-confirmation",
					expectCookieSecureValue: false,
				},
			}

			for name, tc := range testCases {
				t.Run(name, func(t *testing.T) {
					w := httptest.NewRecorder()
					appData := AppData{}

					template := &mockTemplate{}
					template.
						On("Func", w, &aboutPaymentData{App: appData}).
						Return(nil)

					payClient = mockPayClient{BaseURL: tc.baseUrl}

					payClient.
						On("CreatePayment", pay.CreatePaymentBody{
							Amount:      0,
							Reference:   "abc",
							Description: "A payment",
							ReturnUrl:   "/payment-confirmation",
							Email:       "a@b.com",
							Language:    "en",
						}).
						Return(pay.CreatePaymentResponse{
							PaymentId: "a-fake-id",
							Links: map[string]pay.Link{
								"next_url": pay.Link{
									Href: tc.expectedRedirectPath,
								},
							},
						}, nil)

					r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

					sessionsStore := &mockSessionsStore{}

					session := sessions.NewSession(sessionsStore, "pay")

					session.Options = &sessions.Options{
						Path:     "/",
						MaxAge:   5400,
						SameSite: http.SameSiteLaxMode,
						HttpOnly: true,
						Secure:   tc.expectCookieSecureValue,
					}
					session.Values = map[interface{}]interface{}{"paymentId": "a-fake-id"}

					sessionsStore.
						On("Save", r, w, session).
						Return(nil)

					err := AboutPayment(&mockLogger{}, template.Func, sessionsStore, &payClient)(appData, w, r)
					resp := w.Result()

					assert.Nil(t, err)
					assert.Equal(t, http.StatusFound, resp.StatusCode)
					assert.Equal(t, tc.expectedRedirectPath, resp.Header.Get("Location"))

					mock.AssertExpectationsForObjects(t, template)
				})
			}
		})
		t.Run("Returns error when cannot create payment", func(t *testing.T) {
			w := httptest.NewRecorder()
			appData := AppData{}
			template := &mockTemplate{}
			payClient = mockPayClient{BaseURL: "http://base.url"}

			payClient.
				On("CreatePayment", pay.CreatePaymentBody{
					Amount:      0,
					Reference:   "abc",
					Description: "A payment",
					ReturnUrl:   "/payment-confirmation",
					Email:       "a@b.com",
					Language:    "en",
				}).
				Return(pay.CreatePaymentResponse{}, expectedError)

			r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

			sessionsStore := &mockSessionsStore{}

			logger := &mockLogger{}
			logger.
				On("Print", "Error creating payment: "+expectedError.Error())

			err := AboutPayment(logger, template.Func, sessionsStore, &payClient)(appData, w, r)

			assert.Equal(t, expectedError, err, "Expected error was not returned")
			mock.AssertExpectationsForObjects(t, logger)

		})
	})

}
