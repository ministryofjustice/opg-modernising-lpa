package signin

import (
	"encoding/json"
	"net/http"
)

func (c *Client) Discover(endpoint string) error {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var discoverResponse DiscoverResponse
	err = json.NewDecoder(res.Body).Decode(&discoverResponse)
	if err != nil {
		return err
	}

	c.discoverData = discoverResponse
	return nil
}
