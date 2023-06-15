package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestReadHappyForm(t *testing.T) {
	form := url.Values{"happy": {"yes"}}
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	assert.Equal(t, &HappyForm{Happy: "yes"}, ReadHappyForm(r))
}

func TestHappyFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *HappyForm
		errors validation.List
	}{
		"yes": {
			form: &HappyForm{Happy: "yes"},
		},
		"no": {
			form: &HappyForm{Happy: "yes"},
		},
		"other": {
			form:   &HappyForm{Happy: "other"},
			errors: validation.With("happy", validation.SelectError{Label: "a-label"}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate("a-label"))
		})
	}
}
