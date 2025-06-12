package donordata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersonToNotifyFullName(t *testing.T) {
	assert.Equal(t, "First Last", PersonToNotify{FirstNames: "First", LastName: "Last"}.FullName())
}

func TestPersonToNotifyNameHasChanged(t *testing.T) {
	assert.False(t, PersonToNotify{FirstNames: "a", LastName: "b"}.NameHasChanged("a", "b"))
	assert.True(t, PersonToNotify{FirstNames: "a", LastName: "b"}.NameHasChanged("a", ""))
}
