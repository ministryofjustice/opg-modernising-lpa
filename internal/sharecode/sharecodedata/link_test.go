package sharecodedata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLinkHasExpired(t *testing.T) {
	now := time.Date(2000, time.January, 2, 13, 14, 15, 0, time.UTC)

	assert.True(t, Link{CreatedAt: time.Date(1998, time.January, 2, 13, 14, 14, 0, time.UTC)}.HasExpired(now))
	assert.False(t, Link{CreatedAt: time.Date(1998, time.January, 2, 13, 14, 15, 0, time.UTC)}.HasExpired(now))
}
