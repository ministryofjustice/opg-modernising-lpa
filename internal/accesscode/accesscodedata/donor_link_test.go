package accesscodedata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDonorLinkFor(t *testing.T) {
	now := time.Date(2000, time.January, 2, 13, 14, 15, 0, time.UTC)
	link := DonorLink{}

	assert.Equal(t, DonorLink{
		UpdatedAt: now,
		ExpiresAt: time.Date(2000, time.April, 2, 13, 14, 15, 0, time.UTC),
	}, link.For(now))
}
