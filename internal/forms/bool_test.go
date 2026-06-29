package forms

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool_Set(t *testing.T) {
	e := NewBool("a", "A")

	e.Set(true)
	assert.Equal(t, "1", e.Input)

	e.Set(false)
	assert.Equal(t, "", e.Input)
}

func TestBool_ParsePostForm(t *testing.T) {
	type formType struct {
		Form
		B *Bool
	}

	t.Run("WithError", func(t *testing.T) {
		aForm := formType{
			B: NewBool("a", "A").
				Selected().
				WithError(newEmptyError("nope")),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.B.Name: {"  1  "}})

			assert.True(t, aForm.ParsePostForm(req, aForm.B))
			assert.Nil(t, aForm.B.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.B.Name: {"  "}})

			assert.False(t, aForm.ParsePostForm(req, aForm.B))
			assert.Equal(t, newEmptyError("nope"), aForm.B.Error)
		})
	})

	t.Run("Selected", func(t *testing.T) {
		aForm := formType{
			B: NewBool("b", "B").Selected(),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.B.Name: {"  1  "}})

			assert.True(t, aForm.ParsePostForm(req, aForm.B))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "1", aForm.B.Input)
			assert.Equal(t, true, aForm.B.Value)
			assert.Nil(t, aForm.B.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.B.Name: {"  "}})

			assert.False(t, aForm.ParsePostForm(req, aForm.B))
			assert.Equal(t, []Field{aForm.B.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.B.Input)
			assert.Equal(t, false, aForm.B.Value)
			assert.Equal(t, newSelectError("B"), aForm.B.Error)
		})
	})
}
