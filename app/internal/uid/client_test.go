package uid

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var validLpa = &page.Lpa{
	Type: "pfa",
	Donor: actor.Donor{
		FirstNames:  "Jane",
		LastName:    "Smith",
		DateOfBirth: date.New("2000", "1", "2"),
		Address:     place.Address{Postcode: "ABC123"},
	},
}

func TestNew(t *testing.T) {
	client := New("http://base-url.com", &http.Client{})

	assert.Equal(t, "http://base-url.com", client.baseUrl)
	assert.Equal(t, &http.Client{}, client.httpClient)
}

func TestCreateCase(t *testing.T) {
	var endpointCalled string
	var contentTypeSet string
	var requestMethod string
	var requestBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		rBody, _ := io.ReadAll(r.Body)

		endpointCalled = r.URL.String()
		contentTypeSet = r.Header.Get("Content-Type")
		requestMethod = r.Method
		requestBody = string(rBody)

		w.Write([]byte(`{"uid": "M-789Q-P4DF-4UX3"}`))
	}))

	defer server.Close()

	client := New(server.URL, server.Client())
	resp, err := client.CreateCase(validLpa)

	expectedBody := `{"type":"pfa","source":"APPLICANT","donor":{"name":"Jane Smith","dob":"2000-01-02","postcode":"ABC123"}}`

	assert.Equal(t, http.MethodPost, requestMethod)
	assert.Equal(t, "/cases", endpointCalled)
	assert.Equal(t, "application/json", contentTypeSet)
	assert.JSONEq(t, expectedBody, requestBody)

	assert.Nil(t, err)
	assert.Nil(t, resp.BadRequestErrors)
	assert.Equal(t, "M-789Q-P4DF-4UX3", resp.Uid)
}

func TestCreateCaseOnInvalidLpaError(t *testing.T) {
	client := New("/", nil)
	_, err := client.CreateCase(&page.Lpa{})

	assert.Equal(t, errors.New("LPA missing details. Requires Type, Donor name, dob and postcode"), err)
}

func TestCreateCaseOnNewRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	defer server.Close()

	client := New(server.URL+"`invalid-url-format", server.Client())
	_, err := client.CreateCase(&page.Lpa{})

	assert.NotNil(t, err)
}

func TestCreateCaseOnDoRequestError(t *testing.T) {
	expectedError := errors.New("an error")

	httpClient := newMockDoer(t)
	httpClient.
		On("Do", mock.Anything).
		Return(nil, expectedError)

	client := New("/", httpClient)
	_, err := client.CreateCase(validLpa)

	assert.Equal(t, expectedError, err)
}

func TestCreateCaseOnJsonNewDecoderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.Write([]byte(`<not json>`))
	}))

	defer server.Close()

	client := New(server.URL, server.Client())
	_, err := client.CreateCase(validLpa)

	assert.IsType(t, &json.SyntaxError{}, err)
}

func TestValid(t *testing.T) {
	testCases := map[string]*page.Lpa{
		"missing all": {},
		"missing type": {
			Donor: actor.Donor{
				FirstNames:  "Jane",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "1", "2"),
				Address:     place.Address{Postcode: "ABC123"},
			},
		},
		"missing donor fullname": {
			Type: "pfa",
			Donor: actor.Donor{
				DateOfBirth: date.New("2000", "1", "2"),
				Address:     place.Address{Postcode: "ABC123"},
			},
		},
		"missing donor date of birth": {
			Type: "pfa",
			Donor: actor.Donor{
				FirstNames: "Jane",
				LastName:   "Smith",
				Address:    place.Address{Postcode: "ABC123"},
			},
		},
		"missing donor postcode": {
			Type: "pfa",
			Donor: actor.Donor{
				FirstNames:  "Jane",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "1", "2"),
			},
		},
	}

	for name, lpa := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.False(t, Valid(lpa))
		})
	}

	assert.True(t, Valid(validLpa))
}

func TestCreateCaseOnBadRequestResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Write([]byte(`{"code":"INVALID_REQUEST","detail":"string","errors":[{"source":"/donor/dob","detail":"must match format YYYY-MM-DD"}]}`))
	}))

	defer server.Close()

	client := New(server.URL, server.Client())
	resp, err := client.CreateCase(validLpa)

	assert.Equal(t, errors.New("must match format YYYY-MM-DD"), err)
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
			expectedError:  errors.New("must match format YYYY-MM-DD"),
		},
		"400 multiple errors": {
			response:       []byte(`{"code":"INVALID_REQUEST","detail":"string","errors":[{"source":"/donor/dob","detail":"must match format YYYY-MM-DD"},{"source":"/donor/dob","detail":"some other error"}]}`),
			responseHeader: http.StatusBadRequest,
			expectedError:  errors.New("must match format YYYY-MM-DD, some other error"),
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
			resp, err := client.CreateCase(validLpa)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, "", resp.Uid)
		})
	}
}
