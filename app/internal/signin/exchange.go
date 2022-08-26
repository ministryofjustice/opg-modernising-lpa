package signin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type tokenRequestBody struct {
	GrantType           string `json:"grant_type"`
	AuthorizationCode   string `json:"code"`
	RedirectUri         string `json:"redirect_uri"`
	ClientAssertionType string `json:"client_assertion_type"`
	ClientAssertion     string `json:"client_assertion"`
}

type tokenResponseBody struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	IdToken     string `json:"id_token"`
}

func (c *Client) Exchange(code string) (string, error) {
	privateKey, err := c.secretsClient.PrivateKey()
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &jwt.RegisteredClaims{
		Audience:  jwt.ClaimStrings{"https://oidc.integration.account.gov.uk/token"},
		Issuer:    c.clientID,
		Subject:   c.clientID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		ID:        c.randomString(12),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	signedAssertion, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	form := url.Values{
		"client_id":             {c.clientID},
		"grant_type":            {"authorization_code"},
		"redirect_uri":          {c.redirectURL},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {signedAssertion},
		"code":                  {code},
	}

	req, err := http.NewRequest("POST", c.openidConfiguration.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token response was not-OK: %d", res.StatusCode)
	}

	var tokenResponse tokenResponseBody
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("could not read token body: %w", err)
	}

	return tokenResponse.AccessToken, err
}
