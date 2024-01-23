package actor

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemberInviteHasExpired(t *testing.T) {
	testcases := map[bool]time.Time{
		true:  time.Now().Add(-time.Hour * 48),
		false: time.Now().Add(-time.Hour * 47),
	}

	for hasExpired, createdAt := range testcases {
		t.Run(fmt.Sprintf("%v", hasExpired), func(t *testing.T) {
			assert.Equal(t, hasExpired, MemberInvite{CreatedAt: createdAt}.HasExpired())
		})
	}
}
