package donordata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndependentWitnessFullName(t *testing.T) {
	d := IndependentWitness{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", d.FullName())
}

func TestIndependentWitnessNameHasChanged(t *testing.T) {
	assert.False(t, IndependentWitness{FirstNames: "a", LastName: "b"}.NameHasChanged("a", "b"))
	assert.True(t, IndependentWitness{FirstNames: "a", LastName: "b"}.NameHasChanged("a", ""))
}
