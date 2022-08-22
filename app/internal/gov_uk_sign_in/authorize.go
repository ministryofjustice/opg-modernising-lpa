package govuksignin

import (
	"fmt"
	"log"
)

func (c *Client) AuthorizeAndRedirect(redirectURI, clientID, state, nonce, scope string) error {
	authUrl := c.DiscoverData.AuthorizationEndpoint

	q := authUrl.Query()
	q.Set("redirect_uri", redirectURI)
	q.Set("client_id", clientID)
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("scope", scope)
	authUrl.RawQuery = q.Encode()

	// Call out to authorize endpoint
	authorizeUrl := fmt.Sprintf("%s?%s", authUrl.Path, authUrl.RawQuery)
	log.Println(authorizeUrl)
	req, err := c.NewRequest("GET", authorizeUrl, nil)
	if err != nil {
		return err
	}

	_, err = c.httpClient.Do(req)

	return err
}
