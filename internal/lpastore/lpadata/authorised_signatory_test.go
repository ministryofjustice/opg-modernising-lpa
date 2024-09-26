package lpadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorisedSignatoryFullName(t *testing.T) {
	assert.Equal(t, "John Smith", AuthorisedSignatory{FirstNames: "John", LastName: "Smith"}.FullName())
}
