package donordata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrespondentFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Correspondent{FirstNames: "John", LastName: "Smith"}.FullName())
}

func TestCorrespondentNameHasChanged(t *testing.T) {
	assert.False(t, Correspondent{FirstNames: "a", LastName: "b"}.NameHasChanged("a", "b"))
	assert.True(t, Correspondent{FirstNames: "a", LastName: "b"}.NameHasChanged("a", ""))
}
