package signin

import (
	"fmt"
	"net/url"
)

func (c *Client) AuthCodeURL(redirectURI, clientID, state, nonce, scope, signInPublicURL string) string {
	q := url.Values{}
	q.Set("redirect_uri", redirectURI)
	q.Set("client_id", clientID)
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("scope", scope)

	return fmt.Sprintf("%s/authorize?%s", signInPublicURL, q.Encode())
}
