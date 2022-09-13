package pay

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreatePayment(t *testing.T) {
	t.Run("POSTs required body content to expected GOVUK Pay create payment endpoint", func(t *testing.T) {
		body := CreatePaymentBody{
			Amount:      5,
			Reference:   "abc123",
			Description: "A payment",
			ReturnUrl:   "/example/url",
			Email:       "a@example.org",
			Language:    "en",
		}

		created := time.Date(2022, time.January, 2, 0, 0, 0, 0, time.UTC)
		expected := CreatePaymentResponse{
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
			Amount:          5,
			Reference:       "abc123",
			Description:     "A payment",
			ReturnUrl:       "/example/url",
			PaymentId:       "hu20sqlact5260q2nanm0q8u93",
			PaymentProvider: "worldpay",
			ProviderId:      "10987654321",
		}

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()

			b, _ := io.ReadAll(req.Body)
			expectedReqBody := `{"amount": 5,"reference" : "abc123","description": "A payment","return_url": "/example/url","email": "a@example.org","language": "en"}`

			assert.Equal(t, req.URL.String(), "/v1/payments", "URL did not match")
			assert.JSONEq(t, expectedReqBody, string(b), "Body did not match")

			rw.WriteHeader(http.StatusCreated)
			rw.Write([]byte(fmt.Sprintf(`
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
  "amount": 5,
  "reference" : "abc123",
  "description": "A payment",
  "return_url": "/example/url",
  "payment_id": "hu20sqlact5260q2nanm0q8u93",
  "payment_provider": "worldpay",
  "provider_id": "10987654321"
}`, created.Format(time.RFC3339))))
		}))

		defer server.Close()

		payClient, _ := New(server.URL, server.Client())

		actual, err := payClient.CreatePayment(body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expected, actual)
	})
}
