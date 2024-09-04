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
