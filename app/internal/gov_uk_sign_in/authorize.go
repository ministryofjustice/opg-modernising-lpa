package govuksignin

import (
	"fmt"
	"net/http"
)

func (c *Client) AuthorizeAndRedirect(w http.ResponseWriter, r *http.Request, redirectURI, clientID, state, nonce, scope, signInBaseURL string) {
	authUrl := c.DiscoverData.AuthorizationEndpoint

	q := authUrl.Query()
	q.Set("redirect_uri", redirectURI)
	q.Set("client_id", clientID)
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("scope", scope)
	authUrl.RawQuery = q.Encode()

	authorizeUrl := fmt.Sprintf("%s%s?%s", signInBaseURL, authUrl.Path, authUrl.RawQuery)

	http.Redirect(w, r, authorizeUrl, http.StatusFound)
}
