package govuksignin

import (
	"encoding/json"
	"log"

	"github.com/golang-jwt/jwt"
)

type UserInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Phone         string `json:"phone"`
	PhoneVerified bool   `json:"phone_verified"`
	UpdatedAt     int    `json:"updated_at"`
}

func (c *Client) GetUserInfo(jwt *jwt.Token) (UserInfoResponse, error) {
	req, err := c.NewRequest("GET", c.DiscoverData.UserinfoEndpoint.Path, nil)
	if err != nil {
		return UserInfoResponse{}, err
	}

	var bearer = "Bearer " + jwt.Raw
	req.Header.Add("Authorization", bearer)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return UserInfoResponse{}, err
	}

	defer res.Body.Close()
	var userinfoResponse UserInfoResponse

	err = json.NewDecoder(res.Body).Decode(&userinfoResponse)
	log.Println(res.Body)

	return userinfoResponse, err
}
