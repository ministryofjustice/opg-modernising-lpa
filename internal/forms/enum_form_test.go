package forms

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnumForm_Parse(t *testing.T) {
	aForm := NewEnumForm[testEnum]("yesIfGood", testEnumValues{X: "x", Y: "y"})

	t.Run("valid", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.Enum.Name: {" x "}})

		assert.True(t, aForm.Parse(req))
		assert.Empty(t, aForm.Errors)
		assert.Equal(t, "x", aForm.Enum.Input)
		assert.Equal(t, testEnum("x"), aForm.Enum.Value)
		assert.Nil(t, aForm.Enum.Error)
	})

	t.Run("empty", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.Enum.Name: {"  "}})

		assert.False(t, aForm.Parse(req))
		assert.Equal(t, []Field{aForm.Enum.Field}, aForm.Errors)
		assert.Equal(t, "", aForm.Enum.Input)
		assert.Equal(t, testEnum(""), aForm.Enum.Value)
		assert.Equal(t, newSelectError("yesIfGood"), aForm.Enum.Error)
	})

	t.Run("invalid", func(t *testing.T) {
		req := makeRequest(url.Values{aForm.Enum.Name: {" blah "}})

		assert.False(t, aForm.Parse(req))
		assert.Equal(t, []Field{aForm.Enum.Field}, aForm.Errors)
		assert.Equal(t, "blah", aForm.Enum.Input)
		assert.Equal(t, testEnum(""), aForm.Enum.Value)
		assert.Equal(t, newSelectError("yesIfGood"), aForm.Enum.Error)
	})
}
