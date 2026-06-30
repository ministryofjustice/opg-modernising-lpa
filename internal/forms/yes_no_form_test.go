package forms

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYesNoForm_Parse(t *testing.T) {
	aForm := NewYesNoForm("yesIfGood")

	t.Run("yes", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.YesNo.Name: {" yes "}})

		assert.True(t, aForm.Parse(req))
		assert.Empty(t, aForm.Errors)
		assert.Equal(t, "yes", aForm.YesNo.Input)
		assert.Equal(t, Yes, aForm.YesNo.Value)
		assert.Nil(t, aForm.YesNo.Error)
	})

	t.Run("no", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.YesNo.Name: {" no "}})

		assert.True(t, aForm.Parse(req))
		assert.Empty(t, aForm.Errors)
		assert.Equal(t, "no", aForm.YesNo.Input)
		assert.Equal(t, No, aForm.YesNo.Value)
		assert.Nil(t, aForm.YesNo.Error)
	})

	t.Run("unselected", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.YesNo.Name: {" blah "}})

		assert.False(t, aForm.Parse(req))
		assert.Equal(t, []Field{aForm.YesNo.Field}, aForm.Errors)
		assert.Equal(t, "blah", aForm.YesNo.Input)
		assert.Equal(t, YesNoUnknown, aForm.YesNo.Value)
		assert.Equal(t, newSelectError("yesIfGood"), aForm.YesNo.Error)
	})
}
