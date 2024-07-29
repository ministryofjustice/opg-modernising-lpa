package donordata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDonorFullName(t *testing.T) {
	d := Donor{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}
	whiteSpace := Donor{FirstNames: " ", LastName: "  "}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", d.FullName())
	assert.Equal(t, "", whiteSpace.FullName())
}

func TestAuthorisedSignatoryFullName(t *testing.T) {
	d := AuthorisedSignatory{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", d.FullName())
}

func TestIndependentWitnessFullName(t *testing.T) {
	d := IndependentWitness{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", d.FullName())
}
