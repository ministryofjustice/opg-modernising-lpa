package govuksignin

import (
	"encoding/json"
	"io"
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

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Println(bodyString)

	if err != nil {
		return DiscoverResponse{}, err
	}

	return discoverResponse, err
}
