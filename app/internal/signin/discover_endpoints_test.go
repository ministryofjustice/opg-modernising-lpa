package signin

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHttpClient struct {
	mock.Mock
}

func (m *mockHttpClient) Do(r *http.Request) (*http.Response, error) {
	args := m.Called(r)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestDiscoverEndpoints(t *testing.T) {
	dr := DiscoverResponse{
		AuthorizationEndpoint: "http://example.org/authorize",
		TokenEndpoint:         "http://example.org/token",
		Issuer:                "http://example.org",
		UserinfoEndpoint:      "http://example.org/userinfo",
	}
	body, _ := json.Marshal(dr)

	client := &mockHttpClient{}
	client.
		On("Do", mock.MatchedBy(func(r *http.Request) bool {
			return assert.Equal(t, http.MethodGet, r.Method) && assert.Equal(t, "http://base.uri", r.URL.String())
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, nil)

	// Make a Client
	c := NewClient(client, "http://example.org", nil)
	err := c.Discover("http://base.uri")

	assert.Nil(t, err)
	assert.Equal(t, dr, c.DiscoverData)
	mock.AssertExpectationsForObjects(t, client)
}
