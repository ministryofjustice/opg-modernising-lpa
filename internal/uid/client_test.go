package uid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	time "time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var validBody = &CreateCaseRequestBody{
	Type: "property-and-affairs",
	Donor: DonorDetails{
		Name:     "Jane Smith",
		Dob:      date.New("2000", "1", "2"),
		Postcode: "ABC123",
	},
}

var expectedError = errors.New("an error")

type mockCredentialsProvider struct{}

func (m *mockCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     "abc",
		SecretAccessKey: "",
	}, nil
}

func (m *mockCredentialsProvider) IsExpired() bool {
	return false
}

func TestCreateCase(t *testing.T) {
	var endpointCalled string
	var contentTypeSet string
	var requestMethod string
	var requestBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		rBody, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(rBody))

		endpointCalled = r.URL.String()
		contentTypeSet = r.Header.Get("Content-Type")
		requestMethod = r.Method
		requestBody = string(rBody)

		w.Write([]byte(`{"uid": "M-789Q-P4DF-4UX3"}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client())

	uid, err := client.CreateCase(context.Background(), validBody)

	expectedBody := `{"type":"property-and-affairs","source":"APPLICANT","donor":{"name":"Jane Smith","dob":"2000-01-02","postcode":"ABC123"}}`

	assert.Equal(t, http.MethodPost, requestMethod)
	assert.Equal(t, "/cases", endpointCalled)
	assert.Equal(t, "application/json", contentTypeSet)
	assert.JSONEq(t, expectedBody, requestBody)

	assert.Nil(t, err)
	assert.Equal(t, "M-789Q-P4DF-4UX3", uid)
}

func TestCreateCaseOnInvalidBody(t *testing.T) {
	client := &Client{baseURL: "/"}
	_, err := client.CreateCase(context.Background(), &CreateCaseRequestBody{})

	assert.Equal(t, errors.New("CreateCaseRequestBody missing details. Requires Type, Donor name, dob and postcode"), err)
}

func TestCreateCaseOnNewRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	defer server.Close()

	client := New(server.URL+"`invalid-url-format", server.Client())
	_, err := client.CreateCase(context.Background(), validBody)

	assert.NotNil(t, err)
}

func TestCreateCaseOnDoRequestError(t *testing.T) {
	httpClient := newMockDoer(t)
	httpClient.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	client := New("/", httpClient)
	_, err := client.CreateCase(context.Background(), validBody)

	assert.Equal(t, expectedError, err)
}

func TestCreateCaseOnJsonNewDecoderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Write([]byte(`<not json>`))
	}))

	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.CreateCase(context.Background(), validBody)

	assert.IsType(t, &json.SyntaxError{}, err)
}

func TestValid(t *testing.T) {
	testCases := map[string]*CreateCaseRequestBody{
		"missing all": {},
		"missing type": {
			Source: "APPLICANT",
			Donor: DonorDetails{
				Name:     "Jane Smith",
				Dob:      date.New("2000", "1", "2"),
				Postcode: "ABC123",
			},
		},
		"missing donor name": {
			Type:   "property-and-affairs",
			Source: "APPLICANT",
			Donor: DonorDetails{
				Dob:      date.New("2000", "1", "2"),
				Postcode: "ABC123",
			},
		},
		"missing donor date of birth": {
			Type:   "property-and-affairs",
			Source: "APPLICANT",
			Donor: DonorDetails{
				Name:     "Jane Smith",
				Postcode: "ABC123",
			},
		},
		"missing donor postcode": {
			Type:   "property-and-affairs",
			Source: "APPLICANT",
			Donor: DonorDetails{
				Name: "Jane Smith",
				Dob:  date.New("2000", "1", "2"),
			},
		},
	}

	for name, body := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.False(t, body.Valid())
		})
	}

	assert.True(t, validBody.Valid())
}

func TestCreateCaseOnBadRequestResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Write([]byte(`{"code":"INVALID_REQUEST","detail":"string","errors":[{"source":"/donor/dob","detail":"must match format YYYY-MM-DD"}]}`))
	}))

	defer server.Close()

	client := New(server.URL, server.Client())

	uid, err := client.CreateCase(context.Background(), validBody)

	assert.Equal(t, errors.New("error POSTing to UID service: (400) /donor/dob must match format YYYY-MM-DD"), err)
	assert.Equal(t, "", uid)
}

func TestCreateCaseNonSuccessResponses(t *testing.T) {
	testCases := map[string]struct {
		response       []byte
		responseHeader int
		expectedError  error
	}{
		"400 single error": {
			response:       []byte(`{"code":"INVALID_REQUEST","detail":"string","errors":[{"source":"/donor/dob","detail":"must match format YYYY-MM-DD"}]}`),
			responseHeader: http.StatusBadRequest,
			expectedError:  errors.New("error POSTing to UID service: (400) /donor/dob must match format YYYY-MM-DD"),
		},
		"400 multiple errors": {
			response:       []byte(`{"code":"INVALID_REQUEST","detail":"string","errors":[{"source":"/donor/dob","detail":"must match format YYYY-MM-DD"},{"source":"/donor/dob","detail":"some other error"}]}`),
			responseHeader: http.StatusBadRequest,
			expectedError:  errors.New("error POSTing to UID service: (400) /donor/dob must match format YYYY-MM-DD, /donor/dob some other error"),
		},
		"any other > 400 response": {
			response:       []byte(`some body content`),
			responseHeader: http.StatusTeapot,
			expectedError:  errors.New("error POSTing to UID service: (418) some body content"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()

				w.WriteHeader(tc.responseHeader)
				w.Write(tc.response)
			}))

			defer server.Close()

			client := New(server.URL, server.Client())

			uid, err := client.CreateCase(context.Background(), validBody)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, "", uid)
		})
	}
}

func TestPactContract(t *testing.T) {
	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "modernising-lpa",
		Provider: "data-lpa-uid",
		LogDir:   "../../logs",
		PactDir:  "../../pacts",
	})
	assert.Nil(t, err)

	mockProvider.
		AddInteraction().
		Given("The UID service is available").
		UponReceiving("A POST request with valid LPA details").
		WithRequest(http.MethodPost, "/cases", func(b *consumer.V2RequestBuilder) {
			b.
				// Header("Content-Type", matchers.String("application/json")).
				// Header("Authorization", matchers.Regex("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/execute-api/aws4_request, SignedHeaders=content-length;content-type;host;x-amz-date, Signature=98fe2cb1c34c6de900d291351991ba8aa948ca05b7bff969d781edce9b75ee20", "AWS4-HMAC-SHA256 .*")).
				// Header("X-Amz-Date", matchers.String("20000102T000000Z")).
				JSONBody(matchers.Map{
					"type":   matchers.String("property-and-affairs"),
					"source": matchers.String("APPLICANT"),
					"donor": matchers.Like(map[string]any{
						"name":     "Jane Smith",
						"dob":      "2000-01-02",
						"postcode": "ABC123",
					}),
				})
		}).
		WillRespondWith(http.StatusCreated, func(b *consumer.V2ResponseBuilder) {
			// b.Header("Content-Type", matchers.String("application/json"))
			b.JSONBody(matchers.Map{"uid": matchers.Regex("M-789Q-P4DF-4UX3", "M(-[A-Z0-9]{4}){3}")})
		})

	assert.Nil(t, mockProvider.ExecuteTest(t, func(config consumer.MockServerConfig) error {
		baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

		now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }

		cfg := aws.Config{
			Region:      "eu-west-1",
			Credentials: &mockCredentialsProvider{},
		}

		client := New(baseURL, lambda.New(cfg, v4.NewSigner(), http.DefaultClient, now))

		uid, err := client.CreateCase(context.Background(), &CreateCaseRequestBody{
			Type: "property-and-affairs",
			Donor: DonorDetails{
				Name:     "Jane Smith",
				Dob:      date.New("2000", "1", "2"),
				Postcode: "ABC123",
			},
		})

		assert.NotEmpty(t, uid)
		assert.NoError(t, err)

		return nil
	}))
}

func TestCheckHealth(t *testing.T) {
	var endpointCalled string
	var requestMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		rBody, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(rBody))

		endpointCalled = r.URL.String()
		requestMethod = r.Method

		w.Write([]byte(`{"status":"OK"}`))
	}))

	client := New(server.URL, server.Client())

	err := client.CheckHealth(context.Background())

	assert.Equal(t, http.MethodGet, requestMethod)
	assert.Equal(t, "/health", endpointCalled)
	assert.Nil(t, err)
}

func TestCheckHealthOnNewRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	client := New(server.URL+"`invalid-url-format", server.Client())
	err := client.CheckHealth(context.Background())
	assert.NotNil(t, err)
}

func TestCheckHealthOnDoRequestError(t *testing.T) {
	httpClient := newMockDoer(t)
	httpClient.EXPECT().
		Do(mock.Anything).
		Return(nil, expectedError)

	client := New("/", httpClient)
	err := client.CheckHealth(context.Background())
	assert.Equal(t, expectedError, err)
}

func TestCheckHealthWhenNotOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	client := New(server.URL, server.Client())
	err := client.CheckHealth(context.Background())
	assert.NotNil(t, err)
}
