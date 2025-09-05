package accesscodedata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLinkFor(t *testing.T) {
	now := time.Date(1998, time.January, 2, 13, 14, 15, 0, time.UTC)
	link := Link{}

	assert.Equal(t, Link{
		UpdatedAt: now,
		ExpiresAt: time.Date(2000, time.January, 2, 13, 14, 15, 0, time.UTC),
	}, link.For(now))
}
