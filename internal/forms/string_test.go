package forms

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_Set(t *testing.T) {
	s := NewString("a", "A")
	s.Set("  ok  ")

	assert.Equal(t, "ok", s.Input)
}

func TestString_ParsePostForm(t *testing.T) {
	type formType struct {
		Form
		S *String
	}

	t.Run("WithError", func(t *testing.T) {
		aForm := formType{
			S: NewString("s", "S").NotEmpty().WithError(newEmptyError("nope")),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.S.Name: {" something "}})

			assert.True(t, aForm.ParsePostForm(req, aForm.S))
			assert.Nil(t, aForm.S.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.S.Name: {" "}})

			assert.False(t, aForm.ParsePostForm(req, aForm.S))
			assert.Equal(t, newEmptyError("nope"), aForm.S.Error)
		})
	})

	t.Run("NotEmpty", func(t *testing.T) {
		aForm := formType{
			S: NewString("s", "S").NotEmpty(),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.S.Name: {" something "}})

			assert.True(t, aForm.ParsePostForm(req, aForm.S))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "something", aForm.S.Input)
			assert.Equal(t, "something", aForm.S.Value)
			assert.Nil(t, aForm.S.Error)
		})

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.S.Name: {" "}})

			assert.False(t, aForm.ParsePostForm(req, aForm.S))
			assert.Equal(t, []Field{aForm.S.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.S.Input)
			assert.Equal(t, "", aForm.S.Value)
			assert.Equal(t, newEmptyError("S"), aForm.S.Error)
		})
	})

	t.Run("MaxLength", func(t *testing.T) {
		aForm := formType{
			S: NewString("s", "S").NotEmpty().MaxLength(5),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.S.Name: {" 12345 "}})

			assert.True(t, aForm.ParsePostForm(req, aForm.S))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "12345", aForm.S.Input)
			assert.Equal(t, "12345", aForm.S.Value)
			assert.Nil(t, aForm.S.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.S.Name: {" 123456 "}})

			assert.False(t, aForm.ParsePostForm(req, aForm.S))
			assert.Equal(t, []Field{aForm.S.Field}, aForm.Errors)
			assert.Equal(t, "123456", aForm.S.Input)
			assert.Equal(t, "123456", aForm.S.Value)
			assert.Equal(t, newTooLongError("S", 5), aForm.S.Error)
		})
	})
}
