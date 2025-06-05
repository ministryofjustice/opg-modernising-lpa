package donordata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorisedSignatoryFullName(t *testing.T) {
	d := AuthorisedSignatory{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", d.FullName())
}

func TestAuthorisedSignatoryNameHasChanged(t *testing.T) {
	assert.False(t, AuthorisedSignatory{FirstNames: "a", LastName: "b"}.NameHasChanged("a", "b"))
	assert.True(t, AuthorisedSignatory{FirstNames: "a", LastName: "b"}.NameHasChanged("a", ""))
}
