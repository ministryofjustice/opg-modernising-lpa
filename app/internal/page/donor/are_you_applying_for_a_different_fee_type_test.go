package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var publicUrl = "http://example.org"

func TestGetAreYourApplyingForADifferentFeeType(t *testing.T) {
	random := func(int) string { return "123456789012" }

	t.Run("Handles page data", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/are-you-applying-for-a-different-fee-type", nil)

		template := newMockTemplate(t)
		template.
			On("Execute", w, &areYourApplyingForADifferentFeeTypeData{
				App:     testAppData,
				Options: actor.YesNoValues,
			}).
			Return(nil)

		payClient := newMockPayClient(t)

		err := AreYourApplyingForADifferentFeeType(nil, template.Execute, nil, payClient, publicUrl, random)(testAppData, w, r, &page.Lpa{
			CertificateProvider: actor.CertificateProvider{},
		})
		resp := w.Result()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Returns an error when cannot render template", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/are-you-applying-for-a-different-fee-type", nil)

		template := newMockTemplate(t)
		template.
			On("Execute", w, &areYourApplyingForADifferentFeeTypeData{
				App:     testAppData,
				Options: actor.YesNoValues,
			}).
			Return(expectedError)

		err := AreYourApplyingForADifferentFeeType(nil, template.Execute, nil, nil, publicUrl, random)(testAppData, w, r, &page.Lpa{
			CertificateProvider: actor.CertificateProvider{},
		})
		resp := w.Result()

		assert.Equal(t, expectedError, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestPostAreYourApplyingForADifferentFeeType(t *testing.T) {
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
				redirect: page.Paths.PaymentConfirmation.Format("lpa-id"),
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				form := url.Values{
					"different-fee": {actor.No.String()},
				}

				w := httptest.NewRecorder()
				r, _ := http.NewRequest(http.MethodPost, "/are-you-applying-for-a-different-fee-type", strings.NewReader(form.Encode()))
				r.Header.Add("Content-Type", page.FormUrlEncoded)

				template := newMockTemplate(t)
				template.
					On("Execute", w, &areYourApplyingForADifferentFeeTypeData{
						App:     testAppData,
						Options: actor.YesNoValues,
						Form:    &areYourApplyingForADifferentFeeTypeForm{DifferentFee: actor.No},
					}).
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

				err := AreYourApplyingForADifferentFeeType(nil, template.Execute, sessionStore, payClient, publicUrl, random)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}, CertificateProvider: actor.CertificateProvider{}})
				resp := w.Result()

				assert.Nil(t, err)
				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, tc.redirect, resp.Header.Get("Location"))

			})
		}
	})

	t.Run("Returns error when cannot create payment", func(t *testing.T) {
		form := url.Values{
			"different-fee": {actor.No.String()},
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/are-you-applying-for-a-different-fee-type", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", page.FormUrlEncoded)

		template := newMockTemplate(t)

		sessionStore := newMockSessionStore(t)

		logger := newMockLogger(t)
		logger.
			On("Print", "Error creating payment: "+expectedError.Error())

		payClient := newMockPayClient(t)
		payClient.
			On("CreatePayment", mock.Anything).
			Return(pay.CreatePaymentResponse{}, expectedError)

		err := AreYourApplyingForADifferentFeeType(logger, template.Execute, sessionStore, payClient, publicUrl, random)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{}})

		assert.Equal(t, expectedError, err, "Expected error was not returned")
	})

	t.Run("Returns error when cannot save to session", func(t *testing.T) {
		form := url.Values{
			"different-fee": {actor.No.String()},
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/are-you-applying-for-a-different-fee-type", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", page.FormUrlEncoded)

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

		err := AreYourApplyingForADifferentFeeType(logger, template.Execute, sessionStore, payClient, publicUrl, random)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{}})

		assert.Equal(t, expectedError, err, "Expected error was not returned")
	})

	t.Run("Redirects to evidence required when applying for different fee type", func(t *testing.T) {
		form := url.Values{
			"different-fee": {actor.Yes.String()},
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/are-you-applying-for-a-different-fee-type", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", page.FormUrlEncoded)

		template := newMockTemplate(t)
		template.
			On("Execute", w, &areYourApplyingForADifferentFeeTypeData{
				App:     testAppData,
				Options: actor.YesNoValues,
				Form:    &areYourApplyingForADifferentFeeTypeForm{DifferentFee: actor.Yes},
			}).
			Return(nil)

		err := AreYourApplyingForADifferentFeeType(nil, template.Execute, nil, nil, publicUrl, random)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}, CertificateProvider: actor.CertificateProvider{}})
		resp := w.Result()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, page.Paths.EvidenceRequired.Format("lpa-id"), resp.Header.Get("Location"))
	})

	t.Run("Invalid form submission", func(t *testing.T) {
		form := url.Values{
			"different-fee": {""},
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodPost, "/are-you-applying-for-a-different-fee-type", strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", page.FormUrlEncoded)

		validationError := validation.With("different-type", validation.SelectError{Label: "whetherApplyingForDifferentFeeType"})

		template := newMockTemplate(t)
		template.
			On("Execute", w, mock.MatchedBy(func(data *areYourApplyingForADifferentFeeTypeData) bool {
				return assert.Equal(t, validationError, data.Errors)
			})).
			Return(nil)

		err := AreYourApplyingForADifferentFeeType(nil, template.Execute, nil, nil, publicUrl, random)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}, CertificateProvider: actor.CertificateProvider{}})
		resp := w.Result()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAreYourApplyingForADifferentFeeTypeFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *areYourApplyingForADifferentFeeTypeForm
		errors validation.List
	}{
		"valid": {
			form: &areYourApplyingForADifferentFeeTypeForm{},
		},
		"invalid": {
			form: &areYourApplyingForADifferentFeeTypeForm{
				Error: expectedError,
			},
			errors: validation.With("different-type", validation.SelectError{Label: "whetherApplyingForDifferentFeeType"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
