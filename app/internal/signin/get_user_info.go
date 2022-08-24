package signin

import (
	"encoding/json"
	"net/http"
)

type UserInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	UpdatedAt     int    `json:"updated_at"`
}

func (c *Client) GetUserInfo(idToken string) (UserInfoResponse, error) {
	req, err := http.NewRequest("GET", c.DiscoverData.UserinfoEndpoint, nil)
	if err != nil {
		return UserInfoResponse{}, err
	}

	req.Header.Add("Authorization", "Bearer "+idToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return UserInfoResponse{}, err
	}

	defer res.Body.Close()
	var userinfoResponse UserInfoResponse

	err = json.NewDecoder(res.Body).Decode(&userinfoResponse)

	return userinfoResponse, err
}
