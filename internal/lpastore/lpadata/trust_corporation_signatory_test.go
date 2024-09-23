package lpadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustCorporationSignatoryFullName(t *testing.T) {
	assert.Equal(t, "John Smith", TrustCorporationSignatory{FirstNames: "John", LastName: "Smith"}.FullName())
}
