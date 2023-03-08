package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPersonFullName(t *testing.T) {
	p := Donor{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", p.FullName())
}
