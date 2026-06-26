package forms

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeRequest(query url.Values) *http.Request {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(query.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return r
}

func TestParsePostForm(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		type formType struct {
			Form
			S *String
		}

		t.Run("NotEmpty", func(t *testing.T) {
			aForm := formType{
				S: NewString("s", "S").NotEmpty(),
			}

			good := makeRequest(url.Values{aForm.S.Name: {" something "}})
			bad := makeRequest(url.Values{aForm.S.Name: {" "}})

			assert.True(t, aForm.ParsePostForm(good, aForm.S))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "something", aForm.S.Input)
			assert.Equal(t, "something", aForm.S.Value)
			assert.Nil(t, aForm.S.Error)

			assert.False(t, aForm.ParsePostForm(bad, aForm.S))
			assert.Equal(t, []Field{aForm.S.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.S.Input)
			assert.Equal(t, "", aForm.S.Value)
			assert.Equal(t, newEmptyError("S"), aForm.S.Error)
		})

		t.Run("MaxLength", func(t *testing.T) {
			aForm := formType{
				S: NewString("s", "S").NotEmpty().MaxLength(5),
			}

			good := makeRequest(url.Values{aForm.S.Name: {" 12345 "}})
			bad := makeRequest(url.Values{aForm.S.Name: {" 123456 "}})

			assert.True(t, aForm.ParsePostForm(good, aForm.S))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "12345", aForm.S.Input)
			assert.Equal(t, "12345", aForm.S.Value)
			assert.Nil(t, aForm.S.Error)

			assert.False(t, aForm.ParsePostForm(bad, aForm.S))
			assert.Equal(t, []Field{aForm.S.Field}, aForm.Errors)
			assert.Equal(t, "123456", aForm.S.Input)
			assert.Equal(t, "123456", aForm.S.Value)
			assert.Equal(t, newTooLongError("S", 5), aForm.S.Error)
		})
	})
}
