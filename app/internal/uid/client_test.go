package uid

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("an error")

func TestNew(t *testing.T) {
	client := New("http://base-url.com", &http.Client{})

	assert.Equal(t, "http://base-url.com", client.baseUrl)
	assert.Equal(t, &http.Client{}, client.httpClient)
}

func TestCreateCase(t *testing.T) {
	body := `{"type":"pfa","source":"APPLICANT","donor":{"name":"Jane Smith","dob":"2000-1-2","postcode":"ABC123"}}`

	var endpointCalled string
	var contentTypeSet string
	var requestBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		rBody, _ := io.ReadAll(r.Body)

		endpointCalled = r.URL.String()
		contentTypeSet = r.Header.Get("Content-Type")
		requestBody = string(rBody)

		w.Write([]byte(`{"uid": "M-789Q-P4DF-4UX3"}`))
	}))

	defer server.Close()

	client := New(server.URL, server.Client())
	resp, err := client.CreateCase(body)

	expectedRBody := `{"type":"pfa","source":"APPLICANT","donor":{"name":"Jane Smith","dob":"2000-1-2","postcode":"ABC123"}}`

	assert.Equal(t, "/cases", endpointCalled)
	assert.Equal(t, "application/json", contentTypeSet)
	assert.JSONEq(t, expectedRBody, requestBody)

	assert.Nil(t, err)
	assert.Equal(t, `{"uid": "M-789Q-P4DF-4UX3"}`, resp)
}

func TestCreateCaseOnNewRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	defer server.Close()

	client := New(server.URL+"`invalid-url-format", server.Client())
	_, err := client.CreateCase("")

	assert.NotNil(t, err)
}

func TestCreateCaseOnDoRequestError(t *testing.T) {
	httpClient := newMockDoer(t)
	httpClient.
		On("Do", mock.Anything).
		Return(nil, expectedError)

	client := New("/", httpClient)
	_, err := client.CreateCase("aa")

	assert.Equal(t, expectedError, err)
}
