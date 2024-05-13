package telemetry

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlogHandler(t *testing.T) {
	h := NewSlogHandler(slog.NewTextHandler(nil, nil))

	assert.Implements(t, (*slog.Handler)(nil), h)
	assert.IsType(t, (*SlogHandler)(nil), h)
	assert.IsType(t, (*SlogHandler)(nil), h.WithAttrs(nil))
	assert.IsType(t, (*SlogHandler)(nil), h.WithGroup("x"))
}
