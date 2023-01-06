package onelogin

import (
	"encoding/json"
	"net/http"
)

type UserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	UpdatedAt     int    `json:"updated_at"`
}

func (c *Client) UserInfo(idToken string) (UserInfo, error) {
	req, err := http.NewRequest("GET", c.openidConfiguration.UserinfoEndpoint, nil)
	if err != nil {
		return UserInfo{}, err
	}
	req.Header.Add("Authorization", "Bearer "+idToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	defer res.Body.Close()

	var userinfoResponse UserInfo
	err = json.NewDecoder(res.Body).Decode(&userinfoResponse)

	return userinfoResponse, err
}
