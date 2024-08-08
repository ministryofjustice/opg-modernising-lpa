package lpadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttorneyFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Attorney{FirstNames: "John", LastName: "Smith"}.FullName())
}
