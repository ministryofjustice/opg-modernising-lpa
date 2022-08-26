package signin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	IdToken      string `json:"id_token"`
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

	body := tokenRequestBody{
		GrantType:           "authorization_code",
		AuthorizationCode:   code,
		RedirectUri:         c.redirectURL,
		ClientAssertionType: "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		ClientAssertion:     signedAssertion,
	}

	var encodedPostBody bytes.Buffer
	if err := json.NewEncoder(&encodedPostBody).Encode(body); err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.openidConfiguration.TokenEndpoint, &encodedPostBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var tokenResponse tokenResponseBody
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	_, err = jwt.Parse(tokenResponse.IdToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return &privateKey.PublicKey, nil
	})

	return tokenResponse.IdToken, err
}
