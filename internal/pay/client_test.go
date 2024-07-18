package pay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	expectedError = errors.New("err")
	apiToken      = "fake-token"
	ctx           = context.WithValue(context.Background(), "a", "b")
	created, _    = time.Parse(time.RFC3339Nano, "2022-09-29T12:43:46.784Z")
)

func TestCreatePayment(t *testing.T) {
	expectedResponse := &CreatePaymentResponse{
		CreatedDate: created,
		State: State{
			Status:   "created",
			Finished: false,
		},
		Links: map[string]Link{
			"self": {
				Href:   "https://publicapi.payments.service.gov.uk/v1/payments/hu20sqlact5260q2nanm0q8u93",
				Method: "GET",
			},
			"next_url": {
				Href:   "https://www.payments.service.gov.uk/secure/bb0a272c-8eaf-468d-b3xf-ae5e000d2231",
				Method: "GET",
			},
		},
		Amount:          8200,
		Reference:       "abc123",
		Description:     "A payment",
		ReturnURL:       "/example/url",
		PaymentID:       "hu20sqlact5260q2nanm0q8u93",
		PaymentProvider: "worldpay",
		ProviderID:      "10987654321",
	}

	expectedReqBody := `{"amount": 8200, "reference": "abc123", "description": "A payment", "return_url": "/example/url", "email": "a@example.org", "language": "en"}`

	var reqBody []byte
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			if reqBody == nil {
				reqBody, _ = io.ReadAll(req.Body)
			}

			return assert.Equal(t, ctx, req.Context()) &&
				assert.Equal(t, http.MethodPost, req.Method) &&
				assert.Equal(t, req.URL.String(), "http://pay/v1/payments") &&
				assert.Equal(t, req.Header.Get("Authorization"), "Bearer fake-token") &&
				assert.Equal(t, req.Header.Get("Content-Type"), "application/json") &&
				assert.JSONEq(t, expectedReqBody, string(reqBody))
		})).
		Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`
{
	"created_date": "%s",
	"state": {
		"status": "created",
		"finished": false
	},
	"_links": {
		"self": {
			"href": "https://publicapi.payments.service.gov.uk/v1/payments/hu20sqlact5260q2nanm0q8u93",
			"method": "GET"
	 },
		"next_url": {
			"href": "https://www.payments.service.gov.uk/secure/bb0a272c-8eaf-468d-b3xf-ae5e000d2231",
			"method": "GET"
		}
	},
	"amount": 8200,
	"reference" : "abc123",
	"description": "A payment",
	"return_url": "/example/url",
	"payment_id": "hu20sqlact5260q2nanm0q8u93",
	"payment_provider": "worldpay",
	"provider_id": "10987654321"
}`, created.Format(time.RFC3339Nano)))),
		}, nil)

	payClient := New(nil, doer, "http://pay", apiToken)

	actualResponse, err := payClient.CreatePayment(ctx, "lpa-uid", CreatePaymentBody{
		Amount:      8200,
		Reference:   "abc123",
		Description: "A payment",
		ReturnURL:   "/example/url",
		Email:       "a@example.org",
		Language:    "en",
	})
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, actualResponse)
}

func TestCreatePaymentWhenNewRequestErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	defer server.Close()

	payClient := Client{baseURL: server.URL + "`invalid-url-format", apiKey: apiToken, doer: server.Client()}

	_, err := payClient.CreatePayment(ctx, "lpa-uid", CreatePaymentBody{})
	assert.NotNil(t, err)
}

func TestCreatePaymentWhenDoerErrors(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	payClient := Client{doer: doer}

	_, err := payClient.CreatePayment(ctx, "lpa-uid", CreatePaymentBody{})
	assert.Equal(t, expectedError, err)
}

func TestCreatePaymentWhenResponseError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("hey")),
		}, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(ctx, "create payment failed", slog.String("body", "hey"), slog.Int("statusCode", http.StatusBadRequest))

	payClient := Client{doer: doer, logger: logger}

	_, err := payClient.CreatePayment(ctx, "lpa-uid", CreatePaymentBody{})
	assert.Error(t, err)
}

func TestCreatePaymentWhenJsonError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader("hey")),
		}, nil)

	payClient := Client{doer: doer}

	_, err := payClient.CreatePayment(ctx, "lpa-uid", CreatePaymentBody{})
	assert.IsType(t, (*json.SyntaxError)(nil), err)
}

func TestGetPayment(t *testing.T) {
	captured, _ := time.Parse(time.RFC3339Nano, "2022-01-02T12:13:14Z")

	expectedResponse := GetPaymentResponse{
		CreatedDate: created,
		Amount:      8200,
		State: State{
			Status:   "success",
			Finished: true,
		},
		Description: "A payment",
		Reference:   "abc123",
		Language:    "en",
		Email:       "a@example.org",
		CardDetails: CardDetails{
			CardBrand:             "Visa",
			CardType:              "debit",
			LastDigitsCardNumber:  "1234",
			FirstDigitsCardNumber: "123456",
			ExpiryDate:            "04/24",
			CardholderName:        "Sherlock Holmes",
			BillingAddress: BillingAddress{
				Line1:    "221 Baker Street",
				Line2:    "Flat b",
				Postcode: "NW1 6XE",
				City:     "London",
				Country:  "GB",
			},
		},
		PaymentID: "hu20sqlact5260q2nanm0q8u93",
		AuthorisationSummary: AuthorisationSummary{
			ThreeDSecure: ThreeDSecure{
				Required: true,
			},
		},
		RefundSummary: RefundSummary{
			Status:          "available",
			AmountAvailable: 4000,
		},
		SettlementSummary: SettlementSummary{
			CaptureSubmitTime: captured,
			CapturedDate:      date.New("2022", "01", "05"),
			SettledDate:       date.New("2022", "01", "05"),
		},
		DelayedCapture:         false,
		Moto:                   false,
		CorporateCardSurcharge: 250,
		TotalAmount:            4000,
		Fee:                    200,
		NetAmount:              3800,
		PaymentProvider:        "worldpay",
		ProviderID:             "10987654321",
		ReturnURL:              "https://your.service.gov.uk/completed",
	}

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			return assert.Equal(t, ctx, req.Context()) &&
				assert.Equal(t, http.MethodGet, req.Method) &&
				assert.Equal(t, req.URL.String(), "http://pay/v1/payments/payment-id") &&
				assert.Equal(t, req.Header.Get("Authorization"), "Bearer fake-token")
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`
{
	"created_date": "%s",
	"amount": 8200,
	"state": {
		"status": "success",
		"finished": true
	},
	"description": "A payment",
	"reference": "abc123",
	"language": "en",
	"metadata": {
		"ledger_code": "AB100",
		"an_internal_reference_number": 200
	},
	"email": "a@example.org",
	"card_details": {
		"card_brand": "Visa",
		"card_type": "debit",
		"last_digits_card_number": "1234",
		"first_digits_card_number": "123456",
		"expiry_date": "04/24",
		"cardholder_name": "Sherlock Holmes",
		"billing_address": {
				"line1": "221 Baker Street",
				"line2": "Flat b",
				"postcode": "NW1 6XE",
				"city": "London",
				"country": "GB"
		}
	},
	"payment_id": "hu20sqlact5260q2nanm0q8u93",
	"authorisation_summary": {
		"three_d_secure": {
			"required": true
		}
	},
	"refund_summary": {
		"status": "available",
		"amount_available": 4000,
		"amount_submitted": 80
	},
	"settlement_summary": {
		"capture_submit_time": "2022-01-02T12:13:14Z",
		"captured_date": "2022-01-05",
		"settled_date": "2022-01-05"
	},
	"delayed_capture": false,
	"moto": false,
	"corporate_card_surcharge": 250,
	"total_amount": 4000,
	"fee": 200,
	"net_amount": 3800,
	"payment_provider": "worldpay",
	"provider_id": "10987654321",
	"return_url": "https://your.service.gov.uk/completed"
}`,
				created.Format(time.RFC3339Nano)))),
		}, nil)

	payClient := Client{baseURL: "http://pay", apiKey: apiToken, doer: doer}

	response, err := payClient.GetPayment(ctx, "payment-id")
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestGetPaymentWhenNewRequestErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	defer server.Close()

	payClient := Client{baseURL: server.URL + "`invalid-url-format", apiKey: apiToken, doer: server.Client()}

	_, err := payClient.GetPayment(ctx, "payment-id")
	assert.Error(t, err)
}

func TestGetPaymentWhenDoerErrors(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	payClient := Client{doer: doer}

	_, err := payClient.GetPayment(ctx, "payment-id")
	assert.Equal(t, expectedError, err)
}

func TestGetPaymentWhenResponseError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("hey")),
		}, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(ctx, "get payment failed", slog.String("body", "hey"), slog.Int("statusCode", http.StatusBadRequest))

	payClient := Client{doer: doer, logger: logger}

	_, err := payClient.GetPayment(ctx, "payment-id")
	assert.Error(t, err)
}

func TestGetPaymentWhenJsonError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("hey")),
		}, nil)

	payClient := Client{doer: doer}

	_, err := payClient.GetPayment(ctx, "payment-id")
	assert.Error(t, err)
}

func TestCanRedirect(t *testing.T) {
	c := &Client{}
	assert.True(t, c.CanRedirect("https://www.payments.service.gov.uk/whatever?hey"))
	assert.True(t, c.CanRedirect("https://card.payments.service.gov.uk/whatever?hey"))
	assert.False(t, c.CanRedirect("https://card.payments.service.gov.co/whatever?hey"))
	assert.False(t, c.CanRedirect("http://card.payments.service.gov.uk/whatever?hey"))
	assert.False(t, c.CanRedirect("http://bad/https://card.payments.service.gov.uk/whatever?hey"))
}
