package accesscodedata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDonorLinkHasExpired(t *testing.T) {
	now := time.Date(2000, time.April, 2, 13, 14, 15, 0, time.UTC)

	assert.True(t, DonorLink{UpdatedAt: time.Date(2000, time.January, 2, 13, 14, 14, 0, time.UTC)}.HasExpired(now))
	assert.False(t, DonorLink{UpdatedAt: time.Date(2000, time.January, 2, 13, 14, 15, 0, time.UTC)}.HasExpired(now))
}
