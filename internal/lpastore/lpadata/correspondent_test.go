package lpadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrespondentFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Correspondent{FirstNames: "John", LastName: "Smith"}.FullName())
}
