package lpadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersonToNotifyFullName(t *testing.T) {
	assert.Equal(t, "John Smith", PersonToNotify{FirstNames: "John", LastName: "Smith"}.FullName())
}
