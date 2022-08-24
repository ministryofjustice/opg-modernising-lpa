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
	expectedUserInfo := UserInfoResponse{Email: "email@example.com"}

	data, _ := json.Marshal(expectedUserInfo)

	client := &mockHttpClient{}
	client.
		On("Do", mock.MatchedBy(func(r *http.Request) bool {
			return assert.Equal(t, http.MethodGet, r.Method) &&
				assert.Equal(t, "http://user-info", r.URL.String()) &&
				assert.Equal(t, "Bearer hey", r.Header.Get("Authorization"))
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(data)),
		}, nil)

	c := NewClient(client, "http://example.org", nil)
	c.DiscoverData = DiscoverResponse{
		UserinfoEndpoint: "http://user-info",
	}

	userinfo, err := c.GetUserInfo("hey")
	assert.Nil(t, err)
	assert.Equal(t, expectedUserInfo, userinfo)

	mock.AssertExpectationsForObjects(t, client)
}
