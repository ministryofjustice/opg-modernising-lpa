package govuksignin

import (
	"encoding/json"
	"log"
)

func (c *Client) DiscoverEndpoints() (DiscoverResponse, error) {
	req, err := c.NewRequest("GET", "/.well-known/openid-configuration", nil)
	if err != nil {
		return DiscoverResponse{}, err
	}

	res, err := c.httpClient.Do(req)

	if err != nil {
		return DiscoverResponse{}, err
	}

	defer res.Body.Close()

	// Add all endpoints needed for future calls to a struct
	var discoverResponse DiscoverResponse
	err = json.NewDecoder(res.Body).Decode(&discoverResponse)
	log.Println(&res.Body)

	if err != nil {
		return DiscoverResponse{}, err
	}

	return discoverResponse, err
}
