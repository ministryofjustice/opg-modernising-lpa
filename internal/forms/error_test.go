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

func TestDateMissingError_Format(t *testing.T) {
	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("D").
		Return("d")
	localizer.EXPECT().
		T("a").
		Return("ar")
	localizer.EXPECT().
		T("year").
		Return("yar")
	localizer.EXPECT().
		T("and").
		Return("an'")
	localizer.EXPECT().
		Format("errorDateMissing", map[string]any{
			"Label":   "d",
			"Missing": "ar yar",
		}).
		Return("ok")

	error := dateMissingError{Label: "D", MissingYear: true}
	assert.Equal(t, "ok", error.Format(localizer))
}

func TestDateMissingError_FormatMissingTwo(t *testing.T) {
	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("D").
		Return("d")
	localizer.EXPECT().
		T("a").
		Return("ar")
	localizer.EXPECT().
		T("month").
		Return("mon")
	localizer.EXPECT().
		T("day").
		Return("da")
	localizer.EXPECT().
		T("and").
		Return("an'")
	localizer.EXPECT().
		Format("errorDateMissing", map[string]any{
			"Label":   "d",
			"Missing": "ar da an' mon",
		}).
		Return("ok")

	error := dateMissingError{Label: "D", MissingMonth: true, MissingDay: true}
	assert.Equal(t, "ok", error.Format(localizer))
}
