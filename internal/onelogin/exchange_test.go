package onelogin

import (
	"context"
)

var ctx = context.Background()

// func TestExchange(t *testing.T) {
//	privateKey, _ := rsa.GenerateKey(rand.New(rand.NewSource(99)), 2048)
//	jwks := keyfunc.NewGiven(map[string]keyfunc.GivenKey{
//		"myKey": keyfunc.NewGivenHMAC([]byte("my-key")),
//	})

//	token, err := (&jwt.Token{
//		Header: map[string]interface{}{
//			"typ": "JWT",
//			"alg": jwt.SigningMethodHS256.Alg(),
//			"kid": "myKey",
//		},
//		Claims: jwt.MapClaims{
//			"iss":   "http://issuer",
//			"aud":   "client-id",
//			"sub":   "hey",
//			"nonce": "my-nonce",
//		},
//		Method: jwt.SigningMethodHS256,
//	}).SignedString([]byte("my-key"))

//	response := tokenResponseBody{
//		AccessToken: "a",
//		TokenType:   "Bearer",
//		IDToken:     token,
//	}

//	data, _ := json.Marshal(response)

//	secretsClient := newMockSecretsClient(t)
//	secretsClient.
//		On("SecretBytes", ctx, secrets.GovUkOneLoginPrivateKey).
//		Return(pem.EncodeToMemory(
//			&pem.Block{
//				Type:  "RSA PRIVATE KEY",
//				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
//			},
//		), nil)

//	httpClient := newMockHttpClient(t)
//	httpClient.
//		On("Do", mock.MatchedBy(func(r *http.Request) bool {
//			clientAssertion, _ := jwt.Parse(r.FormValue("client_assertion"), func(token *jwt.Token) (interface{}, error) {
//				return &privateKey.PublicKey, nil
//			})

//			claims := clientAssertion.Claims.(jwt.MapClaims)

//			return assert.Equal(t, http.MethodPost, r.Method) &&
//				assert.Equal(t, "http://token", r.URL.String()) &&
//				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type")) &&
//				assert.Equal(t, "client-id", r.FormValue("client_id")) &&
//				assert.Equal(t, "authorization_code", r.FormValue("grant_type")) &&
//				assert.Equal(t, "my-code", r.FormValue("code")) &&
//				assert.Equal(t, "http://redirect", r.FormValue("redirect_uri")) &&
//				assert.Equal(t, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", r.FormValue("client_assertion_type")) &&
//				assert.Equal(t, []interface{}{"https://oidc.integration.account.gov.uk/token"}, claims["aud"]) &&
//				assert.Equal(t, "client-id", claims["iss"]) &&
//				assert.Equal(t, "client-id", claims["sub"]) &&
//				assert.Equal(t, "this-is-random", claims["jti"])
//		})).
//		Return(&http.Response{
//			StatusCode: http.StatusOK,
//			Body:       ioutil.NopCloser(bytes.NewReader(data)),
//		}, nil)

//	client := &Client{
//		httpClient:    httpClient,
//		secretsClient: secretsClient,
//		openidConfiguration: openidConfiguration{
//			Issuer:        "http://issuer",
//			TokenEndpoint: "http://token",
//		},
//		clientID:     "client-id",
//		redirectURL:  "http://redirect",
//		randomString: func(i int) string { return "this-is-random" },
//		jwks:         jwks,
//	}

//	idToken, accessToken, err := client.Exchange(ctx, "my-code", "my-nonce")
//	assert.Nil(t, err)
//	assert.Equal(t, "a", accessToken)
//	assert.Equal(t, token, idToken)
// }

// func TestExchangeWhenPrivateKeyError(t *testing.T) {
//	secretsClient := newMockSecretsClient(t)
//	secretsClient.
//		On("SecretBytes", ctx, secrets.GovUkOneLoginPrivateKey).
//		Return([]byte{}, expectedError)

//	client := &Client{
//		secretsClient: secretsClient,
//	}

//	_, _, err := client.Exchange(ctx, "my-code", "my-nonce")
//	assert.Equal(t, expectedError, err)
// }

// func TestExchangeWhenTokenRequestError(t *testing.T) {
//	privateKey, _ := rsa.GenerateKey(rand.New(rand.NewSource(99)), 2048)

//	secretsClient := newMockSecretsClient(t)
//	secretsClient.
//		On("SecretBytes", ctx, secrets.GovUkOneLoginPrivateKey).
//		Return(pem.EncodeToMemory(
//			&pem.Block{
//				Type:  "RSA PRIVATE KEY",
//				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
//			},
//		), nil)

//	httpClient := newMockHttpClient(t)
//	httpClient.
//		On("Do", mock.Anything).
//		Return(&http.Response{}, expectedError)

//	client := &Client{
//		httpClient:    httpClient,
//		secretsClient: secretsClient,
//		openidConfiguration: openidConfiguration{
//			TokenEndpoint: "http://token",
//		},
//		randomString: func(i int) string { return "this-is-random" },
//	}

//	_, _, err := client.Exchange(ctx, "my-code", "my-nonce")
//	assert.Equal(t, expectedError, err)
// }

// func TestExchangeWhenInvalidToken(t *testing.T) {
//	privateKey, _ := rsa.GenerateKey(rand.New(rand.NewSource(99)), 2048)
//	jwks := keyfunc.NewGiven(map[string]keyfunc.GivenKey{
//		"myKey": keyfunc.NewGivenHMAC([]byte("my-key")),
//	})

//	testCases := map[string]struct {
//		claims jwt.MapClaims
//		key    []byte
//	}{
//		"expired": {
//			claims: jwt.MapClaims{
//				"iss":   "http://issuer",
//				"aud":   "client-id",
//				"nonce": "my-nonce",
//				"exp":   time.Now().Add(-time.Minute).Unix(),
//			},
//			key: []byte("my-key"),
//		},
//		"future issued at": {
//			claims: jwt.MapClaims{
//				"iss":   "http://issuer",
//				"aud":   "client-id",
//				"nonce": "my-nonce",
//				"iat":   time.Now().Add(time.Minute).Unix(),
//			},
//			key: []byte("my-key"),
//		},
//		"missing issuer": {
//			claims: jwt.MapClaims{
//				"aud":   "client-id",
//				"nonce": "my-nonce",
//			},
//			key: []byte("my-key"),
//		},
//		"incorrect issuer": {
//			claims: jwt.MapClaims{
//				"iss":   "http://other",
//				"aud":   "client-id",
//				"nonce": "my-nonce",
//			},
//			key: []byte("my-key"),
//		},
//		"missing audience": {
//			claims: jwt.MapClaims{
//				"iss":   "http://issuer",
//				"nonce": "my-nonce",
//			},
//			key: []byte("my-key"),
//		},
//		"incorrect audience": {
//			claims: jwt.MapClaims{
//				"iss":   "http://issuer",
//				"aud":   "other",
//				"nonce": "my-nonce",
//			},
//			key: []byte("my-key"),
//		},
//		"missing nonce": {
//			claims: jwt.MapClaims{
//				"iss": "http://issuer",
//				"aud": "client-id",
//			},
//			key: []byte("my-key"),
//		},
//		"incorrect nonce": {
//			claims: jwt.MapClaims{
//				"iss":   "http://issuer",
//				"aud":   "client-id",
//				"nonce": "other",
//			},
//			key: []byte("my-key"),
//		},
//		"incorrect signature": {
//			claims: jwt.MapClaims{
//				"iss":   "http://issuer",
//				"aud":   "client-id",
//				"nonce": "my-nonce",
//			},
//			key: []byte("other"),
//		},
//	}

//	for name, tc := range testCases {
//		t.Run(name, func(t *testing.T) {
//			token, err := (&jwt.Token{
//				Header: map[string]interface{}{
//					"typ": "JWT",
//					"alg": jwt.SigningMethodHS256.Alg(),
//					"kid": "myKey",
//				},
//				Claims: tc.claims,
//				Method: jwt.SigningMethodHS256,
//			}).SignedString(tc.key)

//			response := tokenResponseBody{
//				AccessToken: "a",
//				TokenType:   "Bearer",
//				IDToken:     token,
//			}

//			data, _ := json.Marshal(response)

//			secretsClient := newMockSecretsClient(t)
//			secretsClient.
//				On("SecretBytes", ctx, secrets.GovUkOneLoginPrivateKey).
//				Return(pem.EncodeToMemory(
//					&pem.Block{
//						Type:  "RSA PRIVATE KEY",
//						Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
//					},
//				), nil)

//			httpClient := newMockHttpClient(t)
//			httpClient.
//				On("Do", mock.MatchedBy(func(r *http.Request) bool {
//					clientAssertion, _ := jwt.Parse(r.FormValue("client_assertion"), func(token *jwt.Token) (interface{}, error) {
//						return &privateKey.PublicKey, nil
//					})

//					claims := clientAssertion.Claims.(jwt.MapClaims)

//					return assert.Equal(t, http.MethodPost, r.Method) &&
//						assert.Equal(t, "http://token", r.URL.String()) &&
//						assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type")) &&
//						assert.Equal(t, "client-id", r.FormValue("client_id")) &&
//						assert.Equal(t, "authorization_code", r.FormValue("grant_type")) &&
//						assert.Equal(t, "my-code", r.FormValue("code")) &&
//						assert.Equal(t, "http://redirect", r.FormValue("redirect_uri")) &&
//						assert.Equal(t, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", r.FormValue("client_assertion_type")) &&
//						assert.Equal(t, []interface{}{"https://oidc.integration.account.gov.uk/token"}, claims["aud"]) &&
//						assert.Equal(t, "client-id", claims["iss"]) &&
//						assert.Equal(t, "client-id", claims["sub"]) &&
//						assert.Equal(t, "this-is-random", claims["jti"])
//				})).
//				Return(&http.Response{
//					StatusCode: http.StatusOK,
//					Body:       ioutil.NopCloser(bytes.NewReader(data)),
//				}, nil)

//			client := &Client{
//				httpClient:    httpClient,
//				secretsClient: secretsClient,
//				openidConfiguration: openidConfiguration{
//					Issuer:        "http://issuer",
//					TokenEndpoint: "http://token",
//				},
//				clientID:     "client-id",
//				redirectURL:  "http://redirect",
//				randomString: func(i int) string { return "this-is-random" },
//				jwks:         jwks,
//			}

//			_, _, err = client.Exchange(ctx, "my-code", "my-nonce")
//			assert.NotNil(t, err)
//		})
//	}
// }
