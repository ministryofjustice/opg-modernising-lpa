package govuksignin

import (
	"encoding/json"
	"fmt"
	"net/url"
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

	var discoverResponse DiscoverResponse
	err = json.NewDecoder(res.Body).Decode(&discoverResponse)
	if err != nil {
		return DiscoverResponse{}, err
	}

	return discoverResponse, nil
}

func (c *Client) assertEndpointsHostsMatchIssuerHost() error {
	endpoints := []*url.URL{
		&c.DiscoverData.AuthorizationEndpoint,
		&c.DiscoverData.TokenEndpoint,
		&c.DiscoverData.UserinfoEndpoint,
	}

	bu, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("error parsing baseURL: %v", err)
	}

	for _, endpoint := range endpoints {
		if bu.Host != endpoint.Host {
			return fmt.Errorf("Host of URL '%s' does not match issuer. Wanted %s, Got: %s", endpoint.RawPath, c.baseURL, endpoint.Host)
		}
	}

	return nil
}
