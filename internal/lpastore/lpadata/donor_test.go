package lpadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDonorFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Donor{FirstNames: "John", LastName: "Smith"}.FullName())
}
