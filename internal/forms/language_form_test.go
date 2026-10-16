package forms

import (
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
)

func TestLanguageForm_Parse(t *testing.T) {
	aForm := NewLanguageForm("whatYouSpeak")

	t.Run("en", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.Language.Name: {" en "}})

		assert.True(t, aForm.Parse(req))
		assert.Empty(t, aForm.Errors)
		assert.Equal(t, "en", aForm.Language.Input)
		assert.Equal(t, localize.En, aForm.Language.Value)
		assert.Nil(t, aForm.Language.Error)
	})

	t.Run("cy", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.Language.Name: {" cy "}})

		assert.True(t, aForm.Parse(req))
		assert.Empty(t, aForm.Errors)
		assert.Equal(t, "cy", aForm.Language.Input)
		assert.Equal(t, localize.Cy, aForm.Language.Value)
		assert.Nil(t, aForm.Language.Error)
	})

	t.Run("unselected", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.Language.Name: {" blah "}})

		assert.False(t, aForm.Parse(req))
		assert.Equal(t, []Field{aForm.Language.Field}, aForm.Errors)
		assert.Equal(t, "blah", aForm.Language.Input)
		assert.Equal(t, localize.Lang(0), aForm.Language.Value)
		assert.Equal(t, newSelectError("whatYouSpeak"), aForm.Language.Error)
	})
}
