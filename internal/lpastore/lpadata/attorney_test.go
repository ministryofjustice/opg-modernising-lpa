package lpadata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAttorneyFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Attorney{FirstNames: "John", LastName: "Smith"}.FullName())
}

func TestAttorneySigned(t *testing.T) {
	now := time.Now()
	assert.True(t, Attorney{SignedAt: &now}.Signed())

	assert.False(t, Attorney{}.Signed())
	assert.False(t, Attorney{SignedAt: &time.Time{}}.Signed())
}
