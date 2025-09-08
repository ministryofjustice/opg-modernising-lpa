package accesscodedata

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
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

func TestLinkForWhenDonor(t *testing.T) {
	now := time.Date(1998, time.January, 2, 13, 14, 15, 0, time.UTC)
	link := Link{
		PK: dynamo.AccessKey(dynamo.DonorAccessKey("hi")),
	}

	assert.Equal(t, Link{
		PK:        dynamo.AccessKey(dynamo.DonorAccessKey("hi")),
		UpdatedAt: now,
		ExpiresAt: time.Date(1998, time.April, 2, 13, 14, 15, 0, time.UTC),
	}, link.For(now))
}
