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
	"github.com/stretchr/testify/require"
)

func TestDIDClientRefresh(t *testing.T) {
	const body = `{
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

	xbytes, _ := base64.RawURLEncoding.DecodeString("NPGA7cyIKtH1nz2CJIH14s9_CtC93NwdCQcEi-ADvxg")
	ybytes, _ := base64.RawURLEncoding.DecodeString("2cTdmHAmZjighly34lXcxEw50cbKFV7FTOdZKhOG7ps")

	refreshIn, err := client.refresh()
	assert.Nil(t, err)
	assert.Equal(t, time.Hour, refreshIn)

	key, err := client.ForKID("did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936")
	assert.Nil(t, err)
	assert.Equal(t, &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xbytes),
		Y:     new(big.Int).SetBytes(ybytes),
	}, key)
}

func TestDIDClientRefreshWhenDoerError(t *testing.T) {
	doer := newMockDoer(t)
	doer.EXPECT().
		Do(mock.Anything).
		Return(&http.Response{}, expectedError)

	client := &didClient{ctx: context.Background(), http: doer}

	timeout, err := client.refresh()

	assert.Equal(t, time.Duration(0), timeout)
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

	assert.Equal(t, time.Duration(0), timeout)
	require.EqualError(t, err, "unexpected response status code 400 for http://example.com/.well-known/did.json")
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

	assert.Equal(t, time.Duration(0), timeout)
	require.EqualError(t, err, "could not unmarshal public key jwk for did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936: failed to unmarshal JSON Web Key: unsupported key: some (kty)")
}

func TestDIDClientForKIDWhenNoControllerID(t *testing.T) {
	client := &didClient{ctx: context.Background()}

	_, err := client.ForKID("")

	assert.Equal(t, ErrConfigurationMissing, err)
}

func TestDIDClientForKIDWhenMalformedKID(t *testing.T) {
	client := &didClient{ctx: context.Background(), controllerID: "an-id"}

	_, err := client.ForKID("not-a-valid-kid")

	require.EqualError(t, err, "malformed kid missing '#'")
}

func TestDIDClientForKIDWhenUnexpectedControllerID(t *testing.T) {
	client := &didClient{ctx: context.Background(), controllerID: "unexpected-controller-id"}

	_, err := client.ForKID("did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936")

	require.EqualError(t, err, "controller id does not match: unexpected-controller-id != did:web:identity.integration.account.gov.uk")
}

func TestDIDClientForKIDWhenMissingJWKForKID(t *testing.T) {
	client := &didClient{ctx: context.Background(), controllerID: "did:web:identity.integration.account.gov.uk"}

	_, err := client.ForKID("did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936")

	require.EqualError(t, err, "missing jwk for kid did:web:identity.integration.account.gov.uk#c9f8da1c87525bb41653583c2d05274e85805ab7d0abc58376c7128129daa936")
}
