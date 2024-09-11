package onelogin

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	validKID         = "did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936"
	validDIDDocument = `{
	"@context" : [ "https://www.w3.org/ns/did/v1", "https://w3id.org/security/jwk/v1" ],
	"id" : "did:web:identity.integration.account.gov.uk",
	"assertionMethod" : [ {
		"type" : "JsonWebKey",
		"id" : "did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936",
		"controller" : "did:web:identity.integration.account.gov.uk",
		"publicKeyJwk" : {
			"kty" : "EC",
			"crv" : "P-256",
			"x" : "NPGA7cyIKtH1nz2CJIH14s9_CtC93NwdCQcEi-ADvxg",
			"y" : "2cTdmHAmZjighly34lXcxEw50cbKFV7FTOdZKhOG7ps",
			"alg" : "ES256"
		}
	} ]
}`

	validXBytes, _ = base64.RawURLEncoding.DecodeString("NPGA7cyIKtH1nz2CJIH14s9_CtC93NwdCQcEi-ADvxg")
	validYBytes, _ = base64.RawURLEncoding.DecodeString("2cTdmHAmZjighly34lXcxEw50cbKFV7FTOdZKhOG7ps")
	validPublicKey = &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(validXBytes),
		Y:     new(big.Int).SetBytes(validYBytes),
	}
)

func TestGetDIDWhenRefreshAfterCacheControlDuration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Cache-Control": {"max-age=1"},
			},
			Body: io.NopCloser(strings.NewReader(validDIDDocument)),
		}, nil).
		Once()
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(validDIDDocument)),
		}, nil).
		Run(func(req *http.Request) {
			// this needs to run after the doer call is finished, hence this go routine to cancel() later
			go func() {
				time.Sleep(time.Millisecond)
				cancel()
			}()
		}).
		Once()

	logger := newMockLogger(t)

	client := getDID(ctx, logger, doer, "identity-url")
	client.refreshRateLimit = time.Duration(0)

	select {
	case <-ctx.Done():
		key, err := client.ForKID(validKID)
		assert.Nil(t, err)
		assert.NotNil(t, key)
	case <-time.After(5 * time.Second):
		t.Log("test timed out")
		t.Fail()
	}
}

func TestGetDIDWhenRefreshAfterError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil).
		Run(func(req *http.Request) {
			// this needs to run after the doer call is finished, hence this go routine to cancel() later
			go func() {
				time.Sleep(time.Millisecond)
				cancel()
			}()
		}).
		Once()

	logger := newMockLogger(t)
	logger.EXPECT().
		WarnContext(ctx, "problem refreshing did document", mock.Anything)

	client := getDID(ctx, logger, doer, "identity-url")
	client.refreshRateLimit = time.Duration(0)

	select {
	case <-ctx.Done():
		_, err := client.ForKID(validKID)
		assert.Equal(t, ErrConfigurationMissing, err)
	case <-time.After(5 * time.Second):
		t.Log("test timed out")
		t.Fail()
	}
}

func TestGetDIDWhenRefreshForced(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var client *didClient

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(validDIDDocument)),
		}, nil).
		Run(func(req *http.Request) {
			client.requestRefresh()
		}).
		Once()
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(validDIDDocument)),
		}, nil).
		Run(func(req *http.Request) {
			// this needs to run after the doer call is finished, hence this go routine to cancel() later
			go func() {
				time.Sleep(time.Millisecond)
				cancel()
			}()
		}).
		Once()

	logger := newMockLogger(t)

	client = getDID(ctx, logger, doer, "identity-url")
	client.refreshRateLimit = time.Duration(0)

	select {
	case <-ctx.Done():
		key, err := client.ForKID(validKID)
		assert.Nil(t, err)
		assert.NotNil(t, key)
	case <-time.After(5 * time.Second):
		t.Log("test timed out")
		t.Fail()
	}
}

func TestDIDClientRefresh(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			return assert.Equal(t, http.MethodGet, req.Method) &&
				assert.Equal(t, "identity-url/.well-known/did.json", req.URL.String())
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Cache-Control": {"max-age=3600, private"},
			},
			Body: io.NopCloser(strings.NewReader(validDIDDocument)),
		}, nil)

	client := &didClient{ctx: context.Background(), http: doer, identityURL: "identity-url"}

	refreshIn, err := client.refresh()
	assert.Nil(t, err)
	assert.Equal(t, time.Hour, refreshIn)

	key, err := client.ForKID(validKID)
	assert.Nil(t, err)
	assert.Equal(t, validPublicKey, key)
}

func TestDIDClientRefreshWhenDoerError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{}, expectedError)

	client := &didClient{ctx: context.Background(), http: doer}

	timeout, err := client.refresh()

	assert.Equal(t, time.Minute, timeout)
	assert.Error(t, err)
}

func TestDIDClientRefreshWhenUnexpectedStatusCode(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("a body")),
		}, nil)

	client := &didClient{ctx: context.Background(), http: doer, identityURL: "http://example.com"}

	timeout, err := client.refresh()

	assert.Equal(t, time.Minute, timeout)
	assert.EqualError(t, err, "unexpected response status code 400 for http://example.com/.well-known/did.json")
}

func TestDIDClientRefreshWhenCannotUnmarshalPublicKey(t *testing.T) {
	const body = `{
	"@context" : [ "https://www.w3.org/ns/did/v1", "https://w3id.org/security/jwk/v1" ],
	"id" : "did:web:identity.integration.account.gov.uk",
	"assertionMethod" : [ {
		"type" : "JsonWebKey",
		"id" : "did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936",
		"controller" : "did:web:identity.integration.account.gov.uk",
		"publicKeyJwk" : {
			"kty" : "some",
			"crv" : "very",
			"x" : "unexpected",
			"y" : "values",
			"alg" : "here"
		}
	} ]
}`

	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.MatchedBy(func(req *http.Request) bool {
			return true
		})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Cache-Control": {"max-age=3600, private"},
			},
			Body: io.NopCloser(strings.NewReader(body)),
		}, nil)

	client := &didClient{ctx: context.Background(), http: doer}

	timeout, err := client.refresh()

	assert.Equal(t, time.Minute, timeout)
	assert.EqualError(t, err, "could not unmarshal public key jwk for did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936: failed to unmarshal JSON Web Key: unsupported key: some (kty)")
}

func TestDIDClientForKIDWhenNoControllerID(t *testing.T) {
	client := &didClient{ctx: context.Background()}

	_, err := client.ForKID("")

	assert.Equal(t, ErrConfigurationMissing, err)
}

func TestDIDClientForKIDWhenMalformedKID(t *testing.T) {
	client := &didClient{ctx: context.Background(), controllerID: "an-id"}

	_, err := client.ForKID("not-a-valid-kid")

	assert.EqualError(t, err, "malformed kid missing '#'")
}

func TestDIDClientForKIDWhenUnexpectedControllerID(t *testing.T) {
	client := &didClient{ctx: context.Background(), controllerID: "unexpected-controller-id"}

	_, err := client.ForKID("did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936")

	assert.EqualError(t, err, "controller id does not match: unexpected-controller-id != did:web:identity.integration.account.gov.uk")
}

func TestDIDClientForKIDWhenMissingJWKForKID(t *testing.T) {
	client := &didClient{ctx: context.Background(), controllerID: "did:web:identity.integration.account.gov.uk"}

	_, err := client.ForKID("did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936")

	assert.EqualError(t, err, "missing jwk for kid did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936")
}

func TestParseCacheControl(t *testing.T) {
	testcases := map[string]struct {
		value    string
		expected time.Duration
	}{
		"valid": {
			value:    "max-age=300",
			expected: 300 * time.Second,
		},
		"some valid": {
			value:    "ok, max-age=what , max-age=300",
			expected: 300 * time.Second,
		},
		"invalid": {
			value:    "heyyyy",
			expected: refreshInterval,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, parseCacheControl(tc.value))
		})
	}
}
