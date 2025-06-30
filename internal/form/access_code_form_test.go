package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestNewAccessCodeForm(t *testing.T) {
	f := NewAccessCodeForm()

	assert.Equal(t, FieldNames.DonorLastName, f.FieldNames.DonorLastName)
	assert.Equal(t, FieldNames.AccessCode, f.FieldNames.AccessCode)
}

func TestAccessCodeFormRead(t *testing.T) {
	form := url.Values{
		FieldNames.DonorLastName: {" Who? "},
		FieldNames.AccessCode:    {" 12 34-AB CD "},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	f := NewAccessCodeForm()
	f.Read(r)

	assert.Equal(t, f.DonorLastName, "Who?")
	assert.Equal(t, f.AccessCode, "1234ABCD")
	assert.Equal(t, f.AccessCodeRaw, "12 34-AB CD")
}

func TestAccessCodeFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *AccessCodeForm
		errors validation.List
	}{
		"valid": {
			form: &AccessCodeForm{
				DonorLastName: "Smith",
				AccessCode:    "1234ABCD",
			},
		},
		"empty": {
			form: &AccessCodeForm{},
			errors: validation.With(FieldNames.DonorLastName, validation.EnterError{Label: "donorLastName"}).
				With(FieldNames.AccessCode, validation.EnterError{Label: "yourAccessCode"}),
		},
		"wrong length": {
			form: &AccessCodeForm{
				DonorLastName: strings.Repeat("a", 62),
				AccessCode:    "1234",
			},
			errors: validation.With(FieldNames.DonorLastName, validation.StringTooLongError{Label: "donorLastName", Length: 61}).
				With(FieldNames.AccessCode, validation.StringLengthError{Label: "theAccessCodeYouEnter", Length: 8}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
