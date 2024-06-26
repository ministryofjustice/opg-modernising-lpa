package onelogin

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserInfo(t *testing.T) {
	expectedUserInfo := UserInfo{Email: "email@example.com"}

	data, _ := json.Marshal(expectedUserInfo)

	httpClient := newMockDoer(t)
	httpClient.EXPECT().
		Do(mock.MatchedBy(func(r *http.Request) bool {
			return assert.Equal(t, http.MethodGet, r.Method) &&
				assert.Equal(t, "http://user-info", r.URL.String()) &&
				assert.Equal(t, "Bearer hey", r.Header.Get("Authorization"))
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
		}, nil)

	c := &Client{
		httpClient: httpClient,
		openidConfiguration: &configurationClient{
			currentConfiguration: &openidConfiguration{
				UserinfoEndpoint: "http://user-info",
			},
		},
	}

	userinfo, err := c.UserInfo(context.Background(), "hey")
	assert.Nil(t, err)
	assert.Equal(t, expectedUserInfo, userinfo)
}

func TestUserInfoWhenConfigurationError(t *testing.T) {
	c := &Client{
		openidConfiguration: &configurationClient{},
	}

	_, err := c.UserInfo(context.Background(), "hey")
	assert.Equal(t, ErrConfigurationMissing, err)
}

func TestUserInfoWhenRequestError(t *testing.T) {
	httpClient := newMockDoer(t)
	httpClient.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{}, expectedError)

	c := &Client{
		httpClient: httpClient,
		openidConfiguration: &configurationClient{
			currentConfiguration: &openidConfiguration{
				UserinfoEndpoint: "http://user-info",
			},
		},
	}

	_, err := c.UserInfo(context.Background(), "hey")
	assert.Equal(t, expectedError, err)
}

func TestParseIdentityClaim(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	issuedAt := time.Now().Add(-time.Minute).Round(time.Second)

	c := &Client{identityPublicKeyFunc: func(context.Context) (*ecdsa.PublicKey, error) {
		return &privateKey.PublicKey, nil
	}}

	namePart := []map[string]any{
		{
			"validFrom": "2020-03-01",
			"nameParts": []map[string]string{
				{
					"value": "Alice",
					"type":  "GivenName",
				},
				{
					"value": "Jane",
					"type":  "GivenName",
				},
				{
					"value": "Laura",
					"type":  "GivenName",
				},
				{
					"value": "Doe",
					"type":  "FamilyName",
				},
			},
		},
		{
			"validUntil": "2020-03-01",
			"nameParts": []map[string]string{
				{
					"value": "Alice",
					"type":  "GivenName",
				},
				{
					"value": "Eod",
					"type":  "FamilyName",
				},
			},
		},
	}

	birthDatePart := []map[string]any{
		{
			"value": "1970-01-02",
		},
	}

	vc := map[string]any{
		"credentialSubject": map[string]any{
			"name":      namePart,
			"birthDate": birthDatePart,
		},
	}

	mustSign := func(token *jwt.Token, key any) string {
		s, err := token.SignedString(key)

		assert.Nil(t, err)
		return s
	}

	testcases := map[string]struct {
		token    string
		userData identity.UserData
		error    error
	}{
		"with required claims": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc":  vc,
			}), privateKey),
			userData: identity.UserData{
				Status:      identity.StatusConfirmed,
				FirstNames:  "Alice Jane Laura",
				LastName:    "Doe",
				DateOfBirth: date.New("1970", "01", "02"),
				RetrievedAt: issuedAt,
			},
		},
		"missing": {
			error: ErrMissingCoreIdentityJWT,
		},
		"without name": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
			}), privateKey),
			userData: identity.UserData{Status: identity.StatusFailed},
		},
		"without dob": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc": map[string]any{
					"credentialSubject": map[string]any{
						"name": namePart,
					},
				},
			}), privateKey),
			userData: identity.UserData{Status: identity.StatusFailed},
		},
		"with invalid dob": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc": map[string]any{
					"credentialSubject": map[string]any{
						"name": namePart,
						"birthDate": []map[string]any{
							{
								"value": "1970-100-02",
							},
						},
					},
				},
			}), privateKey),
			userData: identity.UserData{Status: identity.StatusFailed},
		},
		"without iat": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"vc": vc,
			}), privateKey),
			userData: identity.UserData{Status: identity.StatusFailed},
		},
		"with unexpected signing method": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"iat": issuedAt.Unix(),
				"vc":  vc,
			}), []byte("a key")),
			error: jwt.ErrTokenUnverifiable,
		},
		"with malformed token": {
			token: "what token",
			error: jwt.ErrTokenMalformed,
		},
		"with invalid token": {
			token: mustSign(jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
				"iat": time.Now().Add(time.Minute).Unix(),
				"vc":  vc,
			}), privateKey),
			error: jwt.ErrTokenInvalidClaims,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			userInfo := UserInfo{
				CoreIdentityJWT: tc.token,
			}

			userData, err := c.ParseIdentityClaim(context.Background(), userInfo)
			assert.ErrorIs(t, err, tc.error)
			assert.Equal(t, tc.userData, userData)
		})
	}
}

func TestParseIdentityClaimWithReturnCode(t *testing.T) {
	testcases := map[string]identity.Status{
		"X":              identity.StatusInsufficientEvidence,
		"any other code": identity.StatusFailed,
	}

	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	c := &Client{identityPublicKeyFunc: func(context.Context) (*ecdsa.PublicKey, error) {
		return &privateKey.PublicKey, nil
	}}

	for returnCode, expectedStatus := range testcases {
		t.Run(returnCode, func(t *testing.T) {
			userInfo := UserInfo{
				ReturnCodes: []ReturnCodeInfo{{Code: returnCode}},
			}

			userData, err := c.ParseIdentityClaim(context.Background(), userInfo)

			assert.Nil(t, err)
			assert.Equal(t, identity.UserData{Status: expectedStatus}, userData)
		})
	}
}

func TestParseIdentityClaimWhenIdentityPublicKeyFuncError(t *testing.T) {
	c := &Client{identityPublicKeyFunc: func(context.Context) (*ecdsa.PublicKey, error) {
		return nil, expectedError
	}}

	_, err := c.ParseIdentityClaim(context.Background(), UserInfo{})
	assert.Equal(t, expectedError, err)
}
