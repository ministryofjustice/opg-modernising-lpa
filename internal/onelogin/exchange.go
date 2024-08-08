package onelogin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
)

type tokenResponseBody struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	IDToken     string `json:"id_token"`
}

func (c *Client) Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error) {
	tokenEndpoint, keyfunc, issuer, err := c.openidConfiguration.ForExchange()
	if err != nil {
		return "", "", err
	}

	privateKeyBytes, err := c.secretsClient.SecretBytes(ctx, secrets.GovUkOneLoginPrivateKey)
	if err != nil {
		return "", "", err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return "", "", err
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
		return "", "", err
	}

	form := url.Values{
		"client_id":             {c.clientID},
		"grant_type":            {"authorization_code"},
		"redirect_uri":          {c.redirectURL},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {signedAssertion},
		"code":                  {code},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("token response was not-OK: %d", res.StatusCode)
	}

	var tokenResponse tokenResponseBody
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return "", "", fmt.Errorf("could not read token body: %w", err)
	}

	if err := c.validateToken(keyfunc, issuer, tokenResponse.IDToken, nonce); err != nil {
		return "", "", fmt.Errorf("id token not valid: %w", err)
	}

	return tokenResponse.IDToken, tokenResponse.AccessToken, err
}

func (c *Client) validateToken(keyfunc jwt.Keyfunc, issuer, idToken, nonce string) error {
	token, err := jwt.ParseWithClaims(idToken, jwt.MapClaims{}, keyfunc,
		jwt.WithIssuer(issuer),
		jwt.WithAudience(c.clientID),
		jwt.WithIssuedAt())
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("idToken not valid")
	}

	if claims := token.Claims.(jwt.MapClaims); claims["nonce"] != nonce {
		return fmt.Errorf("nonce is invalid")
	}

	return nil
}
