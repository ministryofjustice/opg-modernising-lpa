package govuksignin

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

type MockHttpClient struct {
	mock.Mock
}

func (mhc *MockHttpClient) Do(r *http.Request) (*http.Response, error) {
	dr := buildTestDiscoverResponse()
	body, _ := json.Marshal(dr)

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString(string(body))),
	}, nil
}

func buildTestDiscoverResponse() DiscoverResponse {
	ae, _ := url.Parse("http://example.org/authorize")
	te, _ := url.Parse("http://example.org/token")
	issuer := "http://example.org"
	uie, _ := url.Parse("http://example.org/userinfo")

	return DiscoverResponse{
		AuthorizationEndpoint: *ae,
		TokenEndpoint:         *te,
		Issuer:                issuer,
		UserinfoEndpoint:      *uie,
	}
}

func TestDiscoverEndpoints(t *testing.T) {
	mc := MockHttpClient{}

	// Make a Client
	c := NewClient(mc, "http://base.url", "http://example.org")
	got, err := c.DiscoverEndpoints()
	want := buildTestDiscoverResponse()

	assert.Nil(t, err)
	assert.Equal(t, want, got)
}
