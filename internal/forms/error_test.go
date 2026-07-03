package forms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorMessage_Format(t *testing.T) {
	localizer := newMockLocalizer(t)

	localizer.EXPECT().
		T("blah").
		Return("ok")

	error := ErrorMessage("blah")
	assert.Equal(t, "ok", error.Format(localizer))
}

func TestFormattedError_Format(t *testing.T) {
	localizer := newMockLocalizer(t)

	localizer.EXPECT().
		Format("blah", map[string]any{"X": "Y"}).
		Return("ok")

	error := formattedError{
		Key:  "blah",
		Data: map[string]any{"X": "Y"},
	}

	assert.Equal(t, "ok", error.Format(localizer))
}
