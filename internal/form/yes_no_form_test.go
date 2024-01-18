package form

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestReadYesNoForm(t *testing.T) {
	form := url.Values{FieldNames.YesNo: {Yes.String()}}
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	assert.Equal(t, &YesNoForm{YesNo: Yes, ErrorLabel: "a-label", Options: YesNoValues, FieldName: FieldNames.YesNo}, ReadYesNoForm(r, "a-label"))
}

func TestYesNoFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *YesNoForm
		errors validation.List
	}{
		"valid": {
			form: NewYesNoForm(YesNoUnknown),
		},
		"invalid": {
			form:   &YesNoForm{Error: errors.New("err"), ErrorLabel: "a-label", FieldName: FieldNames.YesNo},
			errors: validation.With(FieldNames.YesNo, validation.SelectError{Label: "a-label"}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestNewYesNoForm(t *testing.T) {
	for _, yesNo := range []YesNo{Yes, No, YesNoUnknown} {
		assert.Equal(t, &YesNoForm{
			YesNo:     yesNo,
			Options:   YesNoValues,
			FieldName: "yes-no",
		}, NewYesNoForm(yesNo))
	}
}
