package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActorFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Actor{FirstNames: "John", LastName: "Smith"}.FullName())
}
