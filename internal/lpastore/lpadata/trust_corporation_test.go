package lpadata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTrustCorporationSigned(t *testing.T) {
	assert.True(t, TrustCorporation{Signatories: []TrustCorporationSignatory{
		{SignedAt: time.Now()},
	}}.Signed())

	assert.True(t, TrustCorporation{Signatories: []TrustCorporationSignatory{
		{SignedAt: time.Now()},
		{SignedAt: time.Now()},
	}}.Signed())

	assert.False(t, TrustCorporation{Signatories: []TrustCorporationSignatory{
		{},
	}}.Signed())

	assert.False(t, TrustCorporation{Signatories: []TrustCorporationSignatory{
		{},
		{SignedAt: time.Now()},
	}}.Signed())
}
