package supporterdata

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

func TestMemberInviteFullName(t *testing.T) {
	assert.Equal(t, "a b c", MemberInvite{FirstNames: "a b", LastName: "c"}.FullName())
}

func TestMemberFullName(t *testing.T) {
	assert.Equal(t, "a b c", Member{FirstNames: "a b", LastName: "c"}.FullName())
}
