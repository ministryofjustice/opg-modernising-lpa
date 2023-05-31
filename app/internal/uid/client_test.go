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
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var validBody = &CreateCaseRequestBody{
	Type: "pfa",
	Donor: DonorDetails{
		Name:     "Jane Smith",
		Dob:      ISODate{Time: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
		Postcode: "ABC123",
	},
}

var expectedError = errors.New("an error")

func TestNew(t *testing.T) {
	signer := v4.NewSigner()
	now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }
	client := New("http://base-url.com", "the-region", http.DefaultClient, aws.Credentials{AccessKeyID: "abc"}, signer, now)

	assert.Equal(t, "http://base-url.com", client.baseURL)
	assert.Equal(t, "the-region", client.region)
	assert.Equal(t, http.DefaultClient, client.httpClient)
	assert.Equal(t, aws.Credentials{AccessKeyID: "abc"}, client.credentials)
	assert.Equal(t, signer, client.signer)
	assert.Equal(t, now(), client.now())
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

	now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }

	v4Signer := newMockV4Signer(t)
	v4Signer.
		On("SignHTTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	client := &Client{
		baseURL:     server.URL,
		httpClient:  server.Client(),
		credentials: aws.Credentials{AccessKeyID: "abc"},
		region:      "eu-west-1",
		signer:      v4Signer,
		now:         now,
	}

	resp, err := client.CreateCase(context.Background(), validBody)

	expectedBody := `{"type":"pfa","source":"APPLICANT","donor":{"name":"Jane Smith","dob":"2000-01-02","postcode":"ABC123"}}`

	assert.Equal(t, http.MethodPost, requestMethod)
	assert.Equal(t, "/cases", endpointCalled)
	assert.Equal(t, "application/json", contentTypeSet)
	assert.JSONEq(t, expectedBody, requestBody)

	assert.Nil(t, err)
	assert.Nil(t, resp.BadRequestErrors)
	assert.Equal(t, "M-789Q-P4DF-4UX3", resp.Uid)
}

func TestCreateCaseOnInvalidBody(t *testing.T) {
	client := &Client{baseURL: "/"}
	_, err := client.CreateCase(context.Background(), &CreateCaseRequestBody{})

	assert.Equal(t, errors.New("CreateCaseRequestBody missing details. Requires Type, Donor name, dob and postcode"), err)
}

func TestCreateCaseOnNewRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	defer server.Close()

	client := &Client{
		baseURL:    server.URL + "`invalid-url-format",
		httpClient: server.Client(),
	}
	_, err := client.CreateCase(context.Background(), validBody)

	assert.NotNil(t, err)
}

func TestCreateCaseOnSignError(t *testing.T) {
	now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }

	v4Signer := newMockV4Signer(t)
	v4Signer.
		On("SignHTTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	client := &Client{
		baseURL:     "/",
		httpClient:  nil,
		credentials: aws.Credentials{AccessKeyID: "abc"},
		region:      "eu-west-1",
		signer:      v4Signer,
		now:         now,
	}

	_, err := client.CreateCase(context.Background(), validBody)

	assert.Equal(t, expectedError, err)
}

func TestCreateCaseOnDoRequestError(t *testing.T) {
	v4Signer := newMockV4Signer(t)
	v4Signer.
		On("SignHTTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	httpClient := newMockDoer(t)
	httpClient.
		On("Do", mock.Anything).
		Return(nil, expectedError)

	client := &Client{
		baseURL:     "/",
		httpClient:  httpClient,
		credentials: aws.Credentials{AccessKeyID: "abc"},
		region:      "eu-west-1",
		signer:      v4Signer,
		now:         time.Now,
	}
	_, err := client.CreateCase(context.Background(), validBody)

	assert.Equal(t, expectedError, err)
}

func TestCreateCaseOnJsonNewDecoderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Write([]byte(`<not json>`))
	}))

	defer server.Close()

	v4Signer := newMockV4Signer(t)
	v4Signer.
		On("SignHTTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	client := &Client{
		baseURL:     server.URL,
		httpClient:  server.Client(),
		credentials: aws.Credentials{AccessKeyID: "abc"},
		region:      "eu-west-1",
		signer:      v4Signer,
		now:         time.Now,
	}
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
				Dob:      ISODate{Time: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
				Postcode: "ABC123",
			},
		},
		"missing donor name": {
			Type:   "pfa",
			Source: "APPLICANT",
			Donor: DonorDetails{
				Dob:      ISODate{Time: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
				Postcode: "ABC123",
			},
		},
		"missing donor date of birth": {
			Type:   "pfa",
			Source: "APPLICANT",
			Donor: DonorDetails{
				Name:     "Jane Smith",
				Postcode: "ABC123",
			},
		},
		"missing donor postcode": {
			Type:   "pfa",
			Source: "APPLICANT",
			Donor: DonorDetails{
				Name: "Jane Smith",
				Dob:  ISODate{Time: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
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

	now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }

	v4Signer := newMockV4Signer(t)
	v4Signer.
		On("SignHTTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	client := &Client{
		baseURL:     server.URL,
		httpClient:  server.Client(),
		credentials: aws.Credentials{AccessKeyID: "abc"},
		region:      "eu-west-1",
		signer:      v4Signer,
		now:         now,
	}

	resp, err := client.CreateCase(context.Background(), validBody)

	assert.Equal(t, errors.New("error POSTing to UID service: (400) /donor/dob must match format YYYY-MM-DD"), err)
	assert.Equal(t, "", resp.Uid)
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

			now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }

			v4Signer := newMockV4Signer(t)
			v4Signer.
				On("SignHTTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			client := &Client{
				baseURL:     server.URL,
				httpClient:  server.Client(),
				credentials: aws.Credentials{AccessKeyID: "abc"},
				region:      "eu-west-1",
				signer:      v4Signer,
				now:         now,
			}

			resp, err := client.CreateCase(context.Background(), validBody)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, "", resp.Uid)
		})
	}
}

func TestClientSign(t *testing.T) {
	testCases := map[string]struct {
		Reader        io.Reader
		SignedHeaders string
		Signature     string
	}{
		"empty body": {
			Reader:        nil,
			SignedHeaders: "a-header;host;x-amz-date",
			Signature:     "99f815531e473759852fb13154796d31f4cfaccc3036f91193df440adeba0588",
		},
		"with body": {
			Reader:        strings.NewReader(`{"some": "body data"}`),
			SignedHeaders: "a-header;content-length;host;x-amz-date",
			Signature:     "c9f9b78004a45e947d0fd7ea4eba56a86d3234bed2a7b69240231a1beb8150e9",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "/an-url", tc.Reader)
			req.Header.Set("a-header", "with-a-value")

			now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }
			client := &Client{
				baseURL:     "https://base.url",
				httpClient:  http.DefaultClient,
				credentials: aws.Credentials{AccessKeyID: "abc"},
				region:      "eu-west-1",
				signer:      v4.NewSigner(),
				now:         now,
			}

			err := client.sign(req.Context(), req, "service-name")

			assert.Nil(t, err)

			assert.Equal(t, "20000102T000000Z", req.Header.Get("X-Amz-Date"))
			assert.Equal(t, fmt.Sprintf("AWS4-HMAC-SHA256 Credential=abc/20000102/eu-west-1/service-name/aws4_request, SignedHeaders=%s, Signature=%s", tc.SignedHeaders, tc.Signature), req.Header.Get("Authorization"))
			assert.Equal(t, "with-a-value", req.Header.Get("a-header"))
		})

	}
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 1, expectedError
}

func TestClientSignOnReadAllError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/an-url", errReader{})
	req.Header.Set("Content-Length", "100")

	now := func() time.Time { return time.Now() }
	client := &Client{
		baseURL:     "https://base.url",
		httpClient:  http.DefaultClient,
		credentials: aws.Credentials{AccessKeyID: "abc"},
		region:      "eu-west-1",
		signer:      v4.NewSigner(),
		now:         now,
	}

	err := client.sign(req.Context(), req, "")

	assert.Equal(t, expectedError, err)
}

func TestClientSignOnSignHttpError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/an-url", nil)
	req.Header.Set("a-header", "with-a-value")

	now := func() time.Time { return time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC) }
	v4Signer := newMockV4Signer(t)
	v4Signer.
		On("SignHTTP", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	client := &Client{
		baseURL:     "https://base.url",
		httpClient:  http.DefaultClient,
		credentials: aws.Credentials{AccessKeyID: "abc"},
		region:      "eu-west-1",
		signer:      v4Signer,
		now:         now,
	}

	err := client.sign(req.Context(), req, "")

	assert.Equal(t, expectedError, err)
}
