package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYesNo_Empty(t *testing.T) {
	assert.True(t, YesNoUnknown.Empty())
	assert.False(t, Yes.Empty())
	assert.False(t, No.Empty())
}
