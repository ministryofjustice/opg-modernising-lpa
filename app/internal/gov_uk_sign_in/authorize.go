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

	// Call out to authorize endpoint
	authorizeUrl := fmt.Sprintf("%s%s?%s", signInBaseURL, authUrl.Path, authUrl.RawQuery)

	//req, err := c.NewRequest("GET", authorizeUrl, nil)
	//if err != nil {
	//	log.Fatalf("error building request: %v", err)
	//}

	http.Redirect(w, r, authorizeUrl, http.StatusFound)
}
