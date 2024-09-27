package localize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleMessage(t *testing.T) {
	plain := newSingleMessage("s", nil)
	templated := newSingleMessage("{{.X}}", nil)
	badTemplate := newSingleMessage("{{.X", nil)

	assert.Equal(t, "s", plain.Execute(nil))
	assert.Equal(t, "t", templated.Execute(map[string]any{"X": "t"}))
	assert.Equal(t, "{{.X", badTemplate.Execute(map[string]any{"X": "t"}))
}
