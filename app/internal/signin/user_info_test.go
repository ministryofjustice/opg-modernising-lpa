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

func TestGetUserInfo(t *testing.T) {
	expectedUserInfo := UserInfo{Email: "email@example.com"}

	data, _ := json.Marshal(expectedUserInfo)

	httpClient := &mockHttpClient{}
	httpClient.
		On("Do", mock.MatchedBy(func(r *http.Request) bool {
			return assert.Equal(t, http.MethodGet, r.Method) &&
				assert.Equal(t, "http://user-info", r.URL.String()) &&
				assert.Equal(t, "Bearer hey", r.Header.Get("Authorization"))
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(data)),
		}, nil)

	c := &Client{
		httpClient: httpClient,
		openidConfiguration: openidConfiguration{
			UserinfoEndpoint: "http://user-info",
		},
	}

	userinfo, err := c.UserInfo("hey")
	assert.Nil(t, err)
	assert.Equal(t, expectedUserInfo, userinfo)

	mock.AssertExpectationsForObjects(t, httpClient)
}
