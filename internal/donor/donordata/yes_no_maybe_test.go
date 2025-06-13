package donordata

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestReadYesNoMaybeForm(t *testing.T) {
	form := url.Values{
		"option": {Yes.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	f := ReadYesNoMaybeForm(r, "hey")
	assert.Equal(t, "hey", f.errorLabel)
	assert.Equal(t, Yes, f.Option)
}

func TestYesNoMaybeFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *YesNoMaybeForm
		errors validation.List
	}{
		"yes": {
			form: &YesNoMaybeForm{Option: Yes},
		},
		"no": {
			form: &YesNoMaybeForm{Option: No},
		},
		"maybe": {
			form: &YesNoMaybeForm{Option: Maybe},
		},
		"not selected": {
			form: &YesNoMaybeForm{
				errorLabel: "hey",
			},
			errors: validation.With("option", validation.SelectError{Label: "hey"}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
