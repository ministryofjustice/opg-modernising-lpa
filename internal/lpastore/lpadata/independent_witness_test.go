package lpadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndependentWitnessFullName(t *testing.T) {
	assert.Equal(t, "John Smith", IndependentWitness{FirstNames: "John", LastName: "Smith"}.FullName())
}
