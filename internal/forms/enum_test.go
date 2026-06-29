package forms

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEnum string

func (e testEnum) Empty() bool    { return string(e) == "" }
func (e testEnum) String() string { return string(e) }

func (e *testEnum) UnmarshalText(text []byte) error {
	switch string(text) {
	case "x":
		*e = testEnum("x")
	case "y":
		*e = testEnum("y")
	case "":
		*e = testEnum("")
	default:
		return errors.New("could not parse testEnum")
	}

	return nil
}

type testEnumValues struct {
	X testEnum
	Y testEnum
}

func TestEnum_Set(t *testing.T) {
	e := NewEnum[testEnum]("a", "A", testEnumValues{X: "x"})
	e.Set("y")

	assert.Equal(t, "y", e.Input)
}

func TestEnum_ParsePostForm(t *testing.T) {
	type formType struct {
		Form
		E *Enum[testEnum, testEnumValues, *testEnum]
	}

	t.Run("WithError", func(t *testing.T) {
		aForm := formType{
			E: NewEnum[testEnum]("e", "E", testEnumValues{}).
				Selected().
				WithError(newEmptyError("nope")),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.E.Name: {"  x  "}})

			assert.True(t, aForm.ParsePostForm(req, aForm.E))
			assert.Nil(t, aForm.E.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.E.Name: {"  "}})

			assert.False(t, aForm.ParsePostForm(req, aForm.E))
			assert.Equal(t, newEmptyError("nope"), aForm.E.Error)
		})
	})

	t.Run("Selected", func(t *testing.T) {
		aForm := formType{
			E: NewEnum[testEnum]("e", "E", testEnumValues{}).Selected(),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.E.Name: {"  x  "}})

			assert.True(t, aForm.ParsePostForm(req, aForm.E))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "x", aForm.E.Input)
			assert.Equal(t, testEnum("x"), aForm.E.Value)
			assert.Nil(t, aForm.E.Error)
		})

		t.Run("empty", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.E.Name: {"  "}})

			assert.False(t, aForm.ParsePostForm(req, aForm.E))
			assert.Equal(t, []Field{aForm.E.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.E.Input)
			assert.Equal(t, testEnum(""), aForm.E.Value)
			assert.Equal(t, newSelectError("E"), aForm.E.Error)
		})

		t.Run("error", func(t *testing.T) {
			req := makeRequest(url.Values{aForm.E.Name: {"blah"}})

			assert.False(t, aForm.ParsePostForm(req, aForm.E))
			assert.Equal(t, []Field{aForm.E.Field}, aForm.Errors)
			assert.Equal(t, "blah", aForm.E.Input)
			assert.Equal(t, testEnum(""), aForm.E.Value)
			assert.Equal(t, newSelectError("E"), aForm.E.Error)
		})
	})
}
