package signin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var b64 = base64.URLEncoding.WithPadding(base64.NoPadding)

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

func (c *Client) GetToken(redirectUri, clientID, JTI, code string) (string, error) {
	privateKey, err := c.secretsClient.PrivateKey()
	if err != nil {
		return "", err
	}

	claims := make(jwt.MapClaims)
	claims["aud"] = []string{"https://oidc.integration.account.gov.uk/token"}
	claims["iss"] = clientID
	claims["sub"] = clientID
	claims["exp"] = time.Now().Add(5 * time.Minute).Unix()
	claims["jti"] = JTI
	claims["iat"] = time.Now().Unix()

	signedAssertion, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		return "", err
	}

	body := &tokenRequestBody{
		GrantType:           "authorization_code",
		AuthorizationCode:   code,
		RedirectUri:         redirectUri,
		ClientAssertionType: "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		ClientAssertion:     signedAssertion,
	}

	encodedPostBody := new(bytes.Buffer)
	err = json.NewEncoder(encodedPostBody).Encode(body)

	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.discoverData.TokenEndpoint, encodedPostBody)
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

	err = json.NewDecoder(res.Body).Decode(&tokenResponse)
	if err != nil {
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
